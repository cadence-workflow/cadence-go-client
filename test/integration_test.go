// Copyright (c) 2017-2020 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"go.uber.org/goleak"
	"go.uber.org/zap/zaptest"

	"go.uber.org/cadence"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/interceptors"
	"go.uber.org/cadence/internal"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
)

type IntegrationTestSuite struct {
	*require.Assertions
	suite.Suite
	config       Config
	rpcClient    *rpcClient
	libClient    client.Client
	domainClient client.DomainClient
	activities   *Activities
	workflows    *Workflows
	worker       worker.Worker
	seq          int64
	taskListName string
	tracer       *tracingInterceptorFactory
}

const (
	ctxTimeout                 = 15 * time.Second
	domainName                 = "integration-test-domain"
	domainCacheRefreshInterval = 20 * time.Second
	testContextKey             = "test-context-key"
)

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// waitForTCP waits until target tcp address is available.
func waitForTCP(timeout time.Duration, addr string) error {
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("failed to wait until %s: %v", addr, ctx.Err())
		default:
			conn, err := d.DialContext(ctx, "tcp", addr)
			if err != nil {
				continue
			}
			_ = conn.Close()
			return nil
		}
	}
}

func (ts *IntegrationTestSuite) SetupSuite() {
	ts.Assertions = require.New(ts.T())
	ts.config = newConfig()
	ts.activities = newActivities()
	ts.workflows = &Workflows{tracer: mocktracer.New()}
	ts.Nil(waitForTCP(time.Minute, ts.config.ServiceAddr))
	var err error
	if ts.config.EnableGrpcAdapter {
		ts.rpcClient, err = newGRPCAdapterClient(ts.config.ServiceName, ts.config.ServiceAddr)
	} else {
		ts.rpcClient, err = newRPCClient(ts.config.ServiceName, ts.config.ServiceAddr)
	}
	ts.NoError(err)
	ts.libClient = client.NewClient(ts.rpcClient.Interface, domainName,
		&client.Options{
			Tracer:             ts.workflows.tracer,
			ContextPropagators: []workflow.ContextPropagator{NewStringMapPropagator([]string{testContextKey})},
		})
	ts.domainClient = client.NewDomainClient(ts.rpcClient.Interface, &client.Options{})
	ts.registerDomain()
	internal.StartVersionMetrics(tally.NoopScope)
}

func (ts *IntegrationTestSuite) TearDownSuite() {
	ts.Assertions = require.New(ts.T())
	ts.rpcClient.Close()
	close(internal.StopMetrics)
	// allow the pollers to shut down, and ensure there are no goroutine leaks.
	// this will wait for up to 1 minute for leaks to subside, but exit relatively quickly if possible.
	max := time.After(time.Minute)
	var last error
	for {
		select {
		case <-max:
			if last != nil {
				ts.NoError(last)
				return
			}
			ts.FailNow("leaks timed out but no error, should be impossible")
		case <-time.After(time.Second):
			// https://github.com/uber-go/cadence-client/issues/739
			last = goleak.Find(goleak.IgnoreTopFunction("go.uber.org/cadence/internal.(*coroutineState).initialYield"))
			if last == nil {
				// no leak, done waiting
				return
			}
			// else wait for another check or the timeout (which will record the latest error)
		}
	}
}

func (ts *IntegrationTestSuite) SetupTest() {
	ts.Assertions = require.New(ts.T())
	ts.seq++
	ts.activities.clearInvoked()
	ts.taskListName = fmt.Sprintf("tl-%v", ts.seq)
	ts.tracer = newtracingInterceptorFactory()
}

func (ts *IntegrationTestSuite) BeforeTest(suiteName, testName string) {
	options := worker.Options{
		Tracer:                            ts.workflows.tracer,
		DisableStickyExecution:            ts.config.IsStickyOff,
		Logger:                            zaptest.NewLogger(ts.T()),
		WorkflowInterceptorChainFactories: []interceptors.WorkflowInterceptorFactory{ts.tracer},
		ContextPropagators:                []workflow.ContextPropagator{NewStringMapPropagator([]string{testContextKey})},
	}

	if testName == "TestNonDeterministicWorkflowQuery" || testName == "TestNonDeterministicWorkflowFailPolicy" {
		options.NonDeterministicWorkflowPolicy = worker.NonDeterministicWorkflowPolicyFailWorkflow

		// disable sticky executon so each workflow yield will require rerunning it from beginning
		options.DisableStickyExecution = true
	}

	ts.worker = worker.New(ts.rpcClient.Interface, domainName, ts.taskListName, options)
	ts.registerWorkflowsAndActivities(ts.worker)
	ts.Nil(ts.worker.Start())
}

func (ts *IntegrationTestSuite) TearDownTest() {
	ts.worker.Stop()
}

func (ts *IntegrationTestSuite) TestBasic() {
	var expected []string
	_, err := ts.executeWorkflow("test-basic", ts.workflows.Basic, &expected)
	ts.NoError(err)
	ts.EqualValues(expected, ts.activities.invoked())
	ts.Equal([]string{"ExecuteWorkflow begin", "ExecuteActivity", "ExecuteActivity", "ExecuteWorkflow end"},
		ts.tracer.GetTrace("go.uber.org/cadence/test.(*Workflows).Basic"))
}

func (ts *IntegrationTestSuite) TestActivityRetryOnError() {
	var expected []string
	_, err := ts.executeWorkflow("test-activity-retry-on-error", ts.workflows.ActivityRetryOnError, &expected)
	ts.NoError(err)
	ts.EqualValues(expected, ts.activities.invoked())
}

func (ts *IntegrationTestSuite) TestActivityRetryOnTimeoutStableError() {
	var expected []string
	_, err := ts.executeWorkflow("test-activity-retry-on-timeout-stable-error", ts.workflows.RetryTimeoutStableErrorWorkflow, &expected)
	ts.Nil(err)
}

func (ts *IntegrationTestSuite) TestActivityRetryOptionsChange() {
	var expected []string
	_, err := ts.executeWorkflow("test-activity-retry-options-change", ts.workflows.ActivityRetryOptionsChange, &expected)
	ts.NoError(err)
	ts.EqualValues(expected, ts.activities.invoked())
}

func (ts *IntegrationTestSuite) TestActivityRetryOnStartToCloseTimeout() {
	var expected []string
	_, err := ts.executeWorkflow(
		"test-activity-retry-on-start2close-timeout",
		ts.workflows.ActivityRetryOnTimeout,
		&expected,
		shared.TimeoutTypeStartToClose)

	ts.NoError(err)
	ts.EqualValues(expected, ts.activities.invoked())
}

func (ts *IntegrationTestSuite) TestActivityRetryOnHBTimeout() {
	var expected []string
	_, err := ts.executeWorkflow("test-activity-retry-on-hbtimeout", ts.workflows.ActivityRetryOnHBTimeout, &expected)
	ts.NoError(err)
	ts.EqualValues(expected, ts.activities.invoked())
}

func (ts *IntegrationTestSuite) TestActivityAutoHeartbeat() {
	var expected []string
	_, err := ts.executeWorkflow("test-activity-auto-heartbeat", ts.workflows.ActivityAutoHeartbeat, &expected)
	ts.NoError(err)
	ts.EqualValues(expected, ts.activities.invoked())
}

func (ts *IntegrationTestSuite) TestContinueAsNew() {
	var result int
	_, err := ts.executeWorkflow("test-continueasnew", ts.workflows.ContinueAsNew, &result, 4, ts.taskListName)
	ts.NoError(err)
	ts.Equal(999, result)
}

func (ts *IntegrationTestSuite) TestContinueAsNewCarryOver() {
	var result string
	startOptions := ts.startWorkflowOptions("test-continueasnew-carryover")
	startOptions.Memo = map[string]interface{}{
		"memoKey": "memoVal",
	}
	startOptions.SearchAttributes = map[string]interface{}{
		"CustomKeywordField": "searchAttr",
	}
	_, err := ts.executeWorkflowWithOption(startOptions, ts.workflows.ContinueAsNewWithOptions, &result, 4, ts.taskListName)
	ts.NoError(err)
	ts.Equal("memoVal,searchAttr", result)
}

func (ts *IntegrationTestSuite) TestCancellation() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	run, err := ts.libClient.ExecuteWorkflow(ctx,
		ts.startWorkflowOptions("test-cancellation"), ts.workflows.Basic)
	ts.NoError(err)
	ts.NotNil(run)
	ts.Nil(ts.libClient.CancelWorkflow(ctx, "test-cancellation", run.GetRunID()))
	err = run.Get(ctx, nil)
	ts.Error(err)
	_, ok := err.(*cadence.CanceledError)
	ts.True(ok)
	ts.Truef(client.IsWorkflowError(err), "err from canceled workflows should be a workflow error: %#v", err)
}

func (ts *IntegrationTestSuite) TestStackTraceQuery() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	run, err := ts.libClient.ExecuteWorkflow(ctx,
		ts.startWorkflowOptions("test-stack-trace-query"), ts.workflows.Basic)
	ts.NoError(err)
	value, err := ts.libClient.QueryWorkflow(ctx, "test-stack-trace-query", run.GetRunID(), "__stack_trace")
	ts.NoError(err)
	ts.NotNil(value)
	var trace string
	ts.NoError(value.Get(&trace))
	ts.True(strings.Contains(trace, "go.uber.org/cadence/test.(*Workflows).Basic"))
}

func (ts *IntegrationTestSuite) TestConsistentQuery() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	// this workflow will start a local activity which blocks for long enough
	// to ensure that consistent query must wait in order to satisfy consistency
	wfOpts := ts.startWorkflowOptions("test-consistent-query")
	wfOpts.DecisionTaskStartToCloseTimeout = 5 * time.Second
	run, err := ts.libClient.ExecuteWorkflow(ctx, wfOpts, ts.workflows.ConsistentQueryWorkflow, 3*time.Second)
	ts.Nil(err)
	// Wait for a second to ensure that first decision task gets started and completed before we send signal.
	// Query cannot be run until first decision task has been completed.
	// If signal occurs right after workflow start then WorkflowStarted and Signal events will both be part of the same
	// decision task. So query will be blocked waiting for signal to complete, this is not what we want because it
	// will not exercise the consistent query code path.
	<-time.After(time.Second)
	err = ts.libClient.SignalWorkflow(ctx, "test-consistent-query", run.GetRunID(), consistentQuerySignalCh, "signal-input")
	ts.NoError(err)

	value, err := ts.libClient.QueryWorkflowWithOptions(ctx, &client.QueryWorkflowWithOptionsRequest{
		WorkflowID:            "test-consistent-query",
		RunID:                 run.GetRunID(),
		QueryType:             "consistent_query",
		QueryConsistencyLevel: shared.QueryConsistencyLevelStrong.Ptr(),
	})
	ts.Nil(err)
	ts.NotNil(value)
	ts.NotNil(value.QueryResult)
	ts.Nil(value.QueryRejected)
	var queryResult string
	ts.NoError(value.QueryResult.Get(&queryResult))
	ts.Equal("signal-input", queryResult)
}

func (ts *IntegrationTestSuite) TestWorkflowIDReuseRejectDuplicate() {
	var result string
	_, err := ts.executeWorkflow(
		"test-workflowidreuse-reject-duplicate",
		ts.workflows.IDReusePolicy,
		&result,
		uuid.New(),
		client.WorkflowIDReusePolicyRejectDuplicate,
		false,
		false,
	)
	ts.Error(err)
	gerr, ok := err.(*workflow.GenericError)
	ts.True(ok)
	ts.True(strings.Contains(gerr.Error(), "WorkflowExecutionAlreadyStartedError"))
	ts.Truef(client.IsWorkflowError(err), "already-started child error should be a workflow error: %#v", err)
}

func (ts *IntegrationTestSuite) TestWorkflowIDReuseAllowDuplicateFailedOnly1() {
	var result string
	_, err := ts.executeWorkflow(
		"test-workflowidreuse-reject-duplicate-failed-only1",
		ts.workflows.IDReusePolicy,
		&result,
		uuid.New(),
		client.WorkflowIDReusePolicyAllowDuplicateFailedOnly,
		false,
		false,
	)
	ts.Error(err)
	gerr, ok := err.(*workflow.GenericError)
	ts.True(ok)
	ts.True(strings.Contains(gerr.Error(), "WorkflowExecutionAlreadyStartedError"))
	ts.Truef(client.IsWorkflowError(err), "already-started child error should be a workflow error: %#v", err)
}

func (ts *IntegrationTestSuite) TestWorkflowIDReuseAllowDuplicateFailedOnly2() {
	var result string
	_, err := ts.executeWorkflow(
		"test-workflowidreuse-reject-duplicate-failed-only2",
		ts.workflows.IDReusePolicy,
		&result,
		uuid.New(),
		client.WorkflowIDReusePolicyAllowDuplicateFailedOnly,
		false,
		true,
	)
	ts.NoError(err)
	ts.Equal("WORLD", result)
}

func (ts *IntegrationTestSuite) TestWorkflowIDReuseAllowDuplicate() {
	var result string
	_, err := ts.executeWorkflow(
		"test-workflowidreuse-allow-duplicate",
		ts.workflows.IDReusePolicy,
		&result,
		uuid.New(),
		client.WorkflowIDReusePolicyAllowDuplicate,
		false,
		false,
	)
	ts.NoError(err)
	ts.Equal("HELLOWORLD", result)
}

func (ts *IntegrationTestSuite) TestWorkflowIDReuseErrorViaStartWorkflow() {
	duplicatedWID := "test-workflowidreuse-duplicate-start-error"
	// setup: run any workflow once to consume the ID
	_, err := ts.executeWorkflow(
		duplicatedWID,
		ts.workflows.SimplestWorkflow,
		nil,
	)
	ts.NoError(err, "basic workflow should succeed")

	// a second attempt should fail
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	opts := ts.startWorkflowOptions(duplicatedWID)
	opts.WorkflowIDReusePolicy = client.WorkflowIDReusePolicyRejectDuplicate
	exec, err := ts.libClient.StartWorkflow(ctx, opts, ts.workflows.SimplestWorkflow)
	ts.Nil(exec)
	ts.Error(err)
	ts.IsType(&shared.WorkflowExecutionAlreadyStartedError{}, err, "should be the known already-started error type")
	ts.False(client.IsWorkflowError(err), "start-workflow rejected errors should not be workflow errors")
}

func (ts *IntegrationTestSuite) TestChildWFRetryOnError() {
	_, err := ts.executeWorkflow("test-childwf-retry-on-error", ts.workflows.ChildWorkflowRetryOnError, nil)
	ts.Error(err)
	ts.Truef(client.IsWorkflowError(err), "child error should be a workflow error: %#v", err)
	ts.EqualValues([]string{"toUpper", "toUpper", "toUpper"}, ts.activities.invoked())
}

func (ts *IntegrationTestSuite) TestChildWFRetryOnTimeout() {
	_, err := ts.executeWorkflow("test-childwf-retry-on-timeout", ts.workflows.ChildWorkflowRetryOnTimeout, nil)
	ts.Error(err)
	ts.Truef(client.IsWorkflowError(err), "child-timeout error should be a workflow error: %#v", err)
	ts.EqualValues([]string{"sleep", "sleep", "sleep"}, ts.activities.invoked())
}

func (ts *IntegrationTestSuite) TestChildWFWithMemoAndSearchAttributes() {
	var result string
	_, err := ts.executeWorkflow("test-childwf-success-memo-searchAttr", ts.workflows.ChildWorkflowSuccess, &result)
	ts.NoError(err)
	ts.EqualValues([]string{"getMemoAndSearchAttr"}, ts.activities.invoked())
	ts.Equal("memoVal, searchAttrVal", result)
	ts.Equal([]string{"ExecuteWorkflow begin", "ExecuteChildWorkflow", "ExecuteWorkflow end"},
		ts.tracer.GetTrace("go.uber.org/cadence/test.(*Workflows).ChildWorkflowSuccess"))
}

func (ts *IntegrationTestSuite) TestChildWFWithParentClosePolicyTerminate() {
	var childWorkflowID string
	_, err := ts.executeWorkflow("test-childwf-parent-close-policy", ts.workflows.ChildWorkflowSuccessWithParentClosePolicyTerminate, &childWorkflowID)
	ts.NoError(err)
	// Need to wait for child workflow to finish as well otherwise test becomes flaky
	ts.waitForWorkflowFinish(childWorkflowID, "")
	resp, err := ts.libClient.DescribeWorkflowExecution(context.Background(), childWorkflowID, "")
	ts.NoError(err)
	ts.True(resp.WorkflowExecutionInfo.GetCloseTime() > 0)
}

func (ts *IntegrationTestSuite) TestChildWFWithParentClosePolicyAbandon() {
	var childWorkflowID string
	_, err := ts.executeWorkflow("test-childwf-parent-close-policy", ts.workflows.ChildWorkflowSuccessWithParentClosePolicyAbandon, &childWorkflowID)
	ts.NoError(err)
	resp, err := ts.libClient.DescribeWorkflowExecution(context.Background(), childWorkflowID, "")
	ts.NoError(err)
	ts.Zerof(resp.WorkflowExecutionInfo.GetCloseTime(), "Expected close time to be zero, got %d. Describe response: %#v", resp.WorkflowExecutionInfo.GetCloseTime(), resp)
}

func (ts *IntegrationTestSuite) TestChildWFCancel() {
	var childWorkflowID string
	_, err := ts.executeWorkflow("test-childwf-cancel", ts.workflows.ChildWorkflowCancel, &childWorkflowID)
	ts.NoError(err)
	resp, err := ts.libClient.DescribeWorkflowExecution(context.Background(), childWorkflowID, "")
	ts.NoError(err)
	ts.Equal(shared.WorkflowExecutionCloseStatusCanceled, resp.WorkflowExecutionInfo.GetCloseStatus())
}

func (ts *IntegrationTestSuite) TestActivityCancelUsingReplay() {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(ts.workflows.ActivityCancelRepro, workflow.RegisterOptions{DisableAlreadyRegisteredCheck: true})
	err := replayer.ReplayPartialWorkflowHistoryFromJSONFile(zaptest.NewLogger(ts.T()), "fixtures/activity.cancel.sm.repro.json", 12)
	ts.NoError(err)
}

func (ts *IntegrationTestSuite) TestActivityCancelRepro() {
	var expected []string
	_, err := ts.executeWorkflow("test-activity-cancel-sm", ts.workflows.ActivityCancelRepro, &expected)
	ts.NoError(err)
	ts.EqualValues(expected, ts.activities.invoked())
}

func (ts *IntegrationTestSuite) TestWorkflowWithLocalActivityCtxPropagation() {
	var expected string
	_, err := ts.executeWorkflow("test-wf-local-activity-ctx-prop", ts.workflows.WorkflowWithLocalActivityCtxPropagation, &expected)
	ts.NoError(err)
	ts.EqualValues(expected, "test-data-in-contexttest-data-in-context")
}

func (ts *IntegrationTestSuite) TestLargeQueryResultError() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	run, err := ts.libClient.ExecuteWorkflow(ctx,
		ts.startWorkflowOptions("test-large-query-error"), ts.workflows.LargeQueryResultWorkflow)
	ts.Nil(err)
	value, err := ts.libClient.QueryWorkflow(ctx, "test-large-query-error", run.GetRunID(), "large_query")
	ts.Error(err)
	ts.False(client.IsWorkflowError(err), "query errors should not be workflow errors, as they are request-related")

	queryErr, ok := err.(*shared.QueryFailedError)
	ts.True(ok)
	ts.Equal("query result size (3000000) exceeds limit (2000000)", queryErr.Message)
	ts.Nil(value)
}

func (ts *IntegrationTestSuite) TestInspectActivityInfo() {
	_, err := ts.executeWorkflow("test-activity-info", ts.workflows.InspectActivityInfo, nil)
	ts.Nil(err)
}

func (ts *IntegrationTestSuite) TestInspectLocalActivityInfo() {
	_, err := ts.executeWorkflow("test-local-activity-info", ts.workflows.InspectLocalActivityInfo, nil)
	ts.Nil(err)
}

func (ts *IntegrationTestSuite) TestDomainUpdate() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	name := domainName
	description := "test-description"
	err := ts.domainClient.Update(ctx, &shared.UpdateDomainRequest{
		Name:        &name,
		UpdatedInfo: &shared.UpdateDomainInfo{Description: &description},
	})
	ts.NoError(err)

	domain, err := ts.domainClient.Describe(ctx, name)
	ts.NoError(err)
	ts.Equal(description, *domain.DomainInfo.Description)
}

func (ts *IntegrationTestSuite) TestNonDeterministicWorkflowFailPolicy() {
	_, err := ts.executeWorkflow("test-nondeterminism-failpolicy", ts.workflows.NonDeterminismSimulatorWorkflow, nil)
	var customErr *internal.CustomError
	ok := errors.As(err, &customErr)
	ts.Truef(ok, "expected CustomError but got %T", err)
	ts.Equal("NonDeterministicWorkflowPolicyFailWorkflow", customErr.Reason())
}

func (ts *IntegrationTestSuite) TestNonDeterministicWorkflowQuery() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	run, err := ts.libClient.ExecuteWorkflow(ctx, ts.startWorkflowOptions("test-nondeterministic-query"), ts.workflows.NonDeterminismSimulatorWorkflow)
	ts.Nil(err)
	err = run.Get(ctx, nil)
	var customErr *internal.CustomError
	ok := errors.As(err, &customErr)
	ts.Truef(ok, "expected CustomError but got %T", err)
	ts.Equal("NonDeterministicWorkflowPolicyFailWorkflow", customErr.Reason())

	// query failed workflow should still work
	value, err := ts.libClient.QueryWorkflow(ctx, "test-nondeterministic-query", run.GetRunID(), "__stack_trace")
	ts.NoError(err)
	ts.NotNil(value)
	var trace string
	ts.NoError(value.Get(&trace))
}

func (ts *IntegrationTestSuite) TestOverrideSpanContext() {
	var result map[string]string
	_, err := ts.executeWorkflow("test-override-span-context", ts.workflows.OverrideSpanContext, &result)
	ts.NoError(err)
	ts.Equal("some-value", result["mockpfx-baggage-some-key"])
}

// TestVersionedWorkflowV1 tests that a workflow started on the worker with VersionedWorkflowV1
// can be replayed on worker with VersionedWorkflowV2 and VersionedWorkflowV3,
// but not on VersionedWorkflowV4, VersionedWorkflowV5, VersionedWorkflowV6.
func (ts *IntegrationTestSuite) TestVersionedWorkflowV1() {
	ts.testVersionedWorkflow(testVersionedWorkflowTestCase{
		version: VersionedWorkflowVersionV1,
		compatibleVersions: []VersionedWorkflowVersion{
			VersionedWorkflowVersionV2,
			VersionedWorkflowVersionV3,
		},
		inCompatibleVersions: []VersionedWorkflowVersion{
			VersionedWorkflowVersionV4,
			VersionedWorkflowVersionV5,
			VersionedWorkflowVersionV6,
		},
	})
}

// TestVersionedWorkflowV2 tests that a workflow started on the worker with VersionedWorkflowV2
// can be replayed on worker with VersionedWorkflowV1 and VersionedWorkflowV3,
// but not on VersionedWorkflowV4, VersionedWorkflowV5, VersionedWorkflowV6.
func (ts *IntegrationTestSuite) TestVersionedWorkflowV2() {
	ts.testVersionedWorkflow(testVersionedWorkflowTestCase{
		version: VersionedWorkflowVersionV2,
		compatibleVersions: []VersionedWorkflowVersion{
			VersionedWorkflowVersionV1,
			VersionedWorkflowVersionV3,
		},
		inCompatibleVersions: []VersionedWorkflowVersion{
			VersionedWorkflowVersionV4,
			VersionedWorkflowVersionV5,
			VersionedWorkflowVersionV6,
		},
	})
}

// TestVersionedWorkflowV3 tests that a workflow started on the worker with VersionedWorkflowV3
// can be replayed on worker with VersionedWorkflowV2, VersionedWorkflowV4, VersionedWorkflowV5, VersionedWorkflowV6
// but not on VersionedWorkflowV1
func (ts *IntegrationTestSuite) TestVersionedWorkflowV3() {
	ts.testVersionedWorkflow(testVersionedWorkflowTestCase{
		version: VersionedWorkflowVersionV3,
		compatibleVersions: []VersionedWorkflowVersion{
			VersionedWorkflowVersionV2,
			VersionedWorkflowVersionV4,
			VersionedWorkflowVersionV5,
			VersionedWorkflowVersionV6,
		},
		inCompatibleVersions: []VersionedWorkflowVersion{
			VersionedWorkflowVersionV1,
		},
	})
}

// TestVersionedWorkflowV4 tests that a workflow started on the worker with VersionedWorkflowV4
// can be replayed on worker with VersionedWorkflowV2, VersionedWorkflowV3, VersionedWorkflowV5, VersionedWorkflowV6
// but not on VersionedWorkflowV1
func (ts *IntegrationTestSuite) TestVersionedWorkflowV4() {
	ts.testVersionedWorkflow(testVersionedWorkflowTestCase{
		version: VersionedWorkflowVersionV4,
		compatibleVersions: []VersionedWorkflowVersion{
			VersionedWorkflowVersionV2,
			VersionedWorkflowVersionV3,
			VersionedWorkflowVersionV5,
			VersionedWorkflowVersionV6,
		},
		inCompatibleVersions: []VersionedWorkflowVersion{
			VersionedWorkflowVersionV1,
		},
	})
}

// TestVersionedWorkflowV5 tests that a workflow started on the worker with VersionedWorkflowV5
// can be replayed on worker with VersionedWorkflowV2, VersionedWorkflowV3, VersionedWorkflowV4, VersionedWorkflowV6,
// but not on VersionedWorkflowV1.
func (ts *IntegrationTestSuite) TestVersionedWorkflowV5() {
	ts.testVersionedWorkflow(testVersionedWorkflowTestCase{
		version: VersionedWorkflowVersionV5,
		compatibleVersions: []VersionedWorkflowVersion{
			VersionedWorkflowVersionV2,
			VersionedWorkflowVersionV3,
			VersionedWorkflowVersionV4,
			VersionedWorkflowVersionV6,
		},
		inCompatibleVersions: []VersionedWorkflowVersion{
			VersionedWorkflowVersionV1,
		},
	})
}

// TestVersionedWorkflowV6 tests that a workflow started on the worker with VersionedWorkflowV6
// can be replayed on worker with VersionedWorkflowV5
// but not on VersionedWorkflowV1, VersionedWorkflowV2, VersionedWorkflowV3, VersionedWorkflowV4.
func (ts *IntegrationTestSuite) TestVersionedWorkflowV6() {
	ts.testVersionedWorkflow(testVersionedWorkflowTestCase{
		version: VersionedWorkflowVersionV6,
		compatibleVersions: []VersionedWorkflowVersion{
			VersionedWorkflowVersionV5,
		},
		inCompatibleVersions: []VersionedWorkflowVersion{
			VersionedWorkflowVersionV1,
			VersionedWorkflowVersionV2,
			VersionedWorkflowVersionV3,
			VersionedWorkflowVersionV4,
		},
	})
}

type testVersionedWorkflowTestCase struct {
	version              VersionedWorkflowVersion
	compatibleVersions   []VersionedWorkflowVersion
	inCompatibleVersions []VersionedWorkflowVersion
}

// testVersionedWorkflow tests that a workflow started on the worker with version
// can be replayed on worker with compatibleVersions
// but not on worker with inCompatibleVersions
func (ts *IntegrationTestSuite) testVersionedWorkflow(c testVersionedWorkflowTestCase) {
	SetupWorkerForVersionedWorkflow(c.version, ts.worker)
	wfID := fmt.Sprintf("test-versioned-workflow-v%d", c.version)
	execution, err := ts.executeWorkflow(wfID, VersionedWorkflowName, nil, "arg")
	ts.NoError(err)

	c.compatibleVersions = append(c.compatibleVersions, c.version)

	ts.Require().Equalf(len(c.compatibleVersions)+len(c.inCompatibleVersions), int(MaxVersionedWorkflowVersion),
		"Test case should cover all versions, but got %d compatible (one of them the testing version itself) and %d incompatible versions, that not equal to %d",
		len(c.compatibleVersions),
		len(c.inCompatibleVersions),
		MaxVersionedWorkflowVersion)

	for _, replayedVersion := range c.compatibleVersions {
		err := ts.replayVersionedWorkflow(replayedVersion, execution)
		ts.NoErrorf(err, "Failed to replay on the replayer with VersionedWorkflowV%d", replayedVersion)
	}

	for _, replayedVersion := range c.inCompatibleVersions {
		err := ts.replayVersionedWorkflow(replayedVersion, execution)
		ts.Errorf(err, "Expected to fail replaying the replayer with VersionedWorkflowV%d", replayedVersion)
	}
}

func (ts *IntegrationTestSuite) replayVersionedWorkflow(version VersionedWorkflowVersion, execution *workflow.Execution) error {
	replayer := worker.NewWorkflowReplayer()
	SetupWorkerForVersionedWorkflow(version, replayer)
	return replayer.ReplayWorkflowExecution(context.Background(), ts.rpcClient, zaptest.NewLogger(ts.T()), domainName, *execution)
}

func (ts *IntegrationTestSuite) registerDomain() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	name := domainName
	retention := int32(1)
	err := ts.domainClient.Register(ctx, &shared.RegisterDomainRequest{
		Name:                                   &name,
		WorkflowExecutionRetentionPeriodInDays: &retention,
	})
	if err != nil {
		if _, ok := err.(*shared.DomainAlreadyExistsError); ok {
			return
		}
	}
	ts.NoError(err)
	time.Sleep(domainCacheRefreshInterval) // wait for domain cache refresh on cadence-server
	// bellow is used to guarantee domain is ready
	var dummyReturn string
	_, err = ts.executeWorkflow("test-domain-exist", ts.workflows.SimplestWorkflow, &dummyReturn)
	numOfRetry := 20
	for err != nil && numOfRetry >= 0 {
		if _, ok := err.(*shared.EntityNotExistsError); ok {
			time.Sleep(domainCacheRefreshInterval)
			_, err = ts.executeWorkflow("test-domain-exist", ts.workflows.SimplestWorkflow, &dummyReturn)
		} else {
			break
		}
		numOfRetry--
	}
}

// executeWorkflow executes a given workflow and waits for the result
func (ts *IntegrationTestSuite) executeWorkflow(wfID string, wfFunc interface{}, retValPtr interface{}, args ...interface{}) (*workflow.Execution, error) {
	options := ts.startWorkflowOptions(wfID)
	return ts.executeWorkflowWithOption(options, wfFunc, retValPtr, args...)
}

func (ts *IntegrationTestSuite) executeWorkflowWithOption(options client.StartWorkflowOptions, wfFunc interface{}, retValPtr interface{}, args ...interface{}) (*workflow.Execution, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	span := ts.workflows.tracer.StartSpan("test-workflow")
	defer span.Finish()
	execution, err := ts.libClient.StartWorkflow(ctx, options, wfFunc, args...)
	if err != nil {
		return nil, err
	}
	run := ts.libClient.GetWorkflow(ctx, execution.ID, execution.RunID)
	err = run.Get(ctx, retValPtr)
	logger := zaptest.NewLogger(ts.T())
	if ts.config.Debug {
		iter := ts.libClient.GetWorkflowHistory(ctx, options.ID, run.GetRunID(), false, shared.HistoryEventFilterTypeAllEvent)
		for iter.HasNext() {
			event, err1 := iter.Next()
			if err1 != nil {
				break
			}
			logger.Info(event.String())
		}
	}
	return execution, err
}

func (ts *IntegrationTestSuite) startWorkflowOptions(wfID string) client.StartWorkflowOptions {
	return client.StartWorkflowOptions{
		ID:                              wfID,
		TaskList:                        ts.taskListName,
		ExecutionStartToCloseTimeout:    15 * time.Second,
		DecisionTaskStartToCloseTimeout: time.Second,
		WorkflowIDReusePolicy:           client.WorkflowIDReusePolicyAllowDuplicate,
	}
}

func (ts *IntegrationTestSuite) registerWorkflowsAndActivities(w worker.Worker) {
	ts.workflows.register(w)
	ts.activities.register(w)
}

func (ts *IntegrationTestSuite) waitForWorkflowFinish(wid string, runID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	wfRun := ts.libClient.GetWorkflow(ctx, wid, runID)
	return wfRun.Get(ctx, nil)
}

var _ interceptors.WorkflowInterceptorFactory = (*tracingInterceptorFactory)(nil)

type tracingInterceptorFactory struct {
	sync.Mutex
	// key is workflow id
	instances map[string]*tracingInterceptor
}

func newtracingInterceptorFactory() *tracingInterceptorFactory {
	return &tracingInterceptorFactory{instances: make(map[string]*tracingInterceptor)}
}

func (t *tracingInterceptorFactory) GetTrace(workflowType string) []string {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	if i, ok := t.instances[workflowType]; ok {
		return i.trace
	}
	panic(fmt.Sprintf("Unknown workflowType %v, known types: %v", workflowType, t.instances))
}
func (t *tracingInterceptorFactory) NewInterceptor(info *workflow.Info, next interceptors.WorkflowInterceptor) interceptors.WorkflowInterceptor {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	result := &tracingInterceptor{
		WorkflowInterceptorBase: interceptors.WorkflowInterceptorBase{Next: next},
	}
	t.instances[info.WorkflowType.Name] = result
	return result
}

var _ interceptors.WorkflowInterceptor = (*tracingInterceptor)(nil)

type tracingInterceptor struct {
	interceptors.WorkflowInterceptorBase
	trace []string
}

func (t *tracingInterceptor) ExecuteActivity(ctx workflow.Context, activityType string, args ...interface{}) workflow.Future {
	t.trace = append(t.trace, "ExecuteActivity")
	return t.Next.ExecuteActivity(ctx, activityType, args...)
}

func (t *tracingInterceptor) ExecuteChildWorkflow(ctx workflow.Context, childWorkflowType string, args ...interface{}) workflow.ChildWorkflowFuture {
	t.trace = append(t.trace, "ExecuteChildWorkflow")
	return t.Next.ExecuteChildWorkflow(ctx, childWorkflowType, args...)
}

func (t *tracingInterceptor) ExecuteWorkflow(ctx workflow.Context, workflowType string, args ...interface{}) []interface{} {
	t.trace = append(t.trace, "ExecuteWorkflow begin")
	result := t.Next.ExecuteWorkflow(ctx, workflowType, args...)
	t.trace = append(t.trace, "ExecuteWorkflow end")
	return result
}
