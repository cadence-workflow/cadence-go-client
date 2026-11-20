// Copyright (c) 2017-2021 Uber Technologies Inc.
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

package internal

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/facebookgo/clock"
	"go.uber.org/zap"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shadower"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/util"
)

const (
	statusInitialized int32 = 0
	statusStarted     int32 = 1
	statusStopped     int32 = 2
)

const (
	defaultWaitDurationPerIteration = 5 * time.Minute
)

type (
	// ShadowOptions is used to configure workflow shadowing.
	ShadowOptions struct {
		// Optional: Workflow visibility query for getting workflows that should be replayed
		// if specified, WorkflowTypes, WorkflowStatus, WorkflowStartTimeFilter fields must not be specified.
		// default: empty query, which matches all workflows
		WorkflowQuery string

		// Optional: A list of workflow type names.
		// The list will be used to construct WorkflowQuery. Only workflows with types listed will be replayed.
		// default: empty list, which matches all workflow types
		WorkflowTypes []string

		// Optional: A list of workflow type names that need to be excluded in the query.
		// The list will be used to construct WorkflowQuery.The listed workflow types will be excluded from replay.
		// default: empty list, which matches all workflow types
		ExcludeTypes []string

		// Optional: A list of workflow status.
		// The list will be used to construct WorkflowQuery. Only workflows with status listed will be replayed.
		// accepted values (case-insensitive): OPEN, CLOSED, ALL, COMPLETED, FAILED, CANCELED, TERMINATED, CONTINUED_AS_NEW, TIMED_OUT
		// default: OPEN, which matches only open workflows
		WorkflowStatus []string

		// Optional: Min and Max workflow start timestamp.
		// Timestamps will be used to construct WorkflowQuery. Only workflows started within the time range will be replayed.
		// default: no time filter, which matches all workflow start timestamp
		WorkflowStartTimeFilter TimeFilter

		// Optional: Min and Max workflow close timestamp.
		// Timestamps will be used to construct WorkflowQuery. Only workflows closed within the time range will be replayed. If this filter is being used along with the thee StartTime filter then make sure the Min Close time stamp
		// is within the range of Start timestamp.
		// default: no time filter, which matches all workflow closed timestamp
		WorkflowCloseTimeFilter TimeFilter

		// Optional: sampling rate for the workflows matches WorkflowQuery
		// only sampled workflows will be replayed
		// default: 1.0
		SamplingRate float64

		// Optional: sets if shadowing should continue after all workflows matches the WorkflowQuery have been replayed.
		// If set to ShadowModeContinuous, ExitCondition must be specified.
		// default: ShadowModeNormal, which means shadowing will complete after all workflows have been replayed
		Mode ShadowMode

		// Required if Mode is set to ShadowModeContinuous: controls when shadowing should complete
		ExitCondition ShadowExitCondition

		// Optional: workflow shadowing concurrency (# of concurrent workflow replay activities)
		// Note: this field only applies to shadow worker. For the local WorkflowShadower,
		// the concurrency will always be 1.
		// An error will be returned if it's set to be larger than 1 when used to NewWorkflowShadower
		// default: 1
		Concurrency int
	}

	// TimeFilter represents a time range through the min and max timestamp
	TimeFilter struct {
		MinTimestamp time.Time
		MaxTimestamp time.Time
	}

	// ShadowMode is an enum for configuring if shadowing should continue after all workflows matches the WorkflowQuery have been replayed.
	ShadowMode int

	// ShadowExitCondition configures when the workflow shadower should exit.
	// If not specified shadower will exit after replaying all workflows satisfying the visibility query.
	ShadowExitCondition struct {
		// Optional: Expiration interval for shadowing.
		// Shadowing will exit when this interval has passed.
		// default: no expiration interval
		ExpirationInterval time.Duration
		// Optional: Target number of shadowed workflows.
		// Shadowing will exit after this number is reached.
		// default: no limit on shadow count
		ShadowCount int
	}

	// WorkflowShadower retrieves and replays workflow history from Cadence service
	// to determine if there's any nondeterministic changes in the workflow definition
	WorkflowShadower struct {
		service       workflowserviceclient.Interface
		domain        string
		shadowOptions ShadowOptions
		logger        *zap.Logger
		replayer      *WorkflowReplayer

		status     int32
		shutdownCh chan struct{}
		shutdownWG sync.WaitGroup

		clock clock.Clock
	}
)

const (
	// ShadowModeNormal is the default mode for workflow shadowing.
	// Shadowing will complete after all workflows matches WorkflowQuery have been replayed.
	ShadowModeNormal ShadowMode = iota
	// ShadowModeContinuous mode will start a new round of shadowing
	// after all workflows matches WorkflowQuery have been replayed.
	// There will be a 5 min wait period between each round,
	// currently this wait period is not configurable.
	// Shadowing will complete only when ExitCondition is met.
	// ExitCondition must be specified when using this mode
	ShadowModeContinuous
)

// NewWorkflowShadower creates an instance of the WorkflowShadower for testing
// The logger is an optional parameter. Defaults to noop logger if not provided and will override the logger in WorkerOptions
func NewWorkflowShadower(
	service workflowserviceclient.Interface,
	domain string,
	shadowOptions ShadowOptions,
	replayOptions ReplayOptions,
	logger *zap.Logger,
) (*WorkflowShadower, error) {
	if len(domain) == 0 {
		return nil, errors.New("domain is not set")
	}

	if err := shadowOptions.validateAndPopulateFields(); err != nil {
		return nil, err
	}

	if shadowOptions.Concurrency > 1 {
		return nil, errors.New("local workflow shadower doesn't support concurrency > 1")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &WorkflowShadower{
		service:       service,
		domain:        domain,
		shadowOptions: shadowOptions,
		logger:        logger,
		replayer:      NewWorkflowReplayerWithOptions(replayOptions),

		status:     statusInitialized,
		shutdownCh: make(chan struct{}),

		clock: clock.New(),
	}, nil
}

// RegisterWorkflow registers workflow function to replay
func (s *WorkflowShadower) RegisterWorkflow(w interface{}) {
	s.replayer.RegisterWorkflow(w)
}

// RegisterWorkflowWithOptions registers workflow function with custom workflow name to replay
func (s *WorkflowShadower) RegisterWorkflowWithOptions(w interface{}, options RegisterWorkflowOptions) {
	s.replayer.RegisterWorkflowWithOptions(w, options)
}

// Run starts WorkflowShadower in a blocking fashion
func (s *WorkflowShadower) Run() error {
	if !atomic.CompareAndSwapInt32(&s.status, statusInitialized, statusStarted) {
		return errors.New("Workflow shadower already started")
	}

	return s.shadowWorker()
}

// Stop stops WorkflowShadower and wait up to one minute for all goroutines to finish before returning
func (s *WorkflowShadower) Stop() {
	if !atomic.CompareAndSwapInt32(&s.status, statusStarted, statusStopped) {
		return
	}

	close(s.shutdownCh)

	if success := util.AwaitWaitGroup(&s.shutdownWG, time.Minute); !success {
		s.logger.Warn("Workflow Shadower timedout on shutdown")
	}
}

// GetRegisteredWorkflows retrieves the list of workflows registered on the worker
func (s *WorkflowShadower) GetRegisteredWorkflows() []RegistryWorkflowInfo {
	return s.replayer.GetRegisteredWorkflows()
}

func (s *WorkflowShadower) shadowWorker() error {
	s.shutdownWG.Add(1)
	defer s.shutdownWG.Done()

	scanRequest := shadower.ScanWorkflowActivityParams{
		Domain:        common.StringPtr(s.domain),
		WorkflowQuery: common.StringPtr(s.shadowOptions.WorkflowQuery),
		SamplingRate:  common.Float64Ptr(s.shadowOptions.SamplingRate),
	}
	s.logger.Info("Shadow workflow query",
		zap.String(tagVisibilityQuery, s.shadowOptions.WorkflowQuery),
	)

	ctx := context.Background()
	expirationTime := time.Unix(0, math.MaxInt64)
	if s.shadowOptions.ExitCondition.ExpirationInterval != 0 {
		expirationTime = s.clock.Now().Add(s.shadowOptions.ExitCondition.ExpirationInterval)
	}

	replayCount := 0
	maxReplayCount := math.MaxInt64
	if s.shadowOptions.ExitCondition.ShadowCount != 0 {
		maxReplayCount = s.shadowOptions.ExitCondition.ShadowCount
	}
	rand.Seed(s.clock.Now().UnixNano())
	for {
		scanResult, err := scanWorkflowExecutionsHelper(ctx, s.service, scanRequest, s.logger)
		if err != nil {
			return err
		}

		for _, execution := range scanResult.Executions {
			if s.clock.Now().After(expirationTime) {
				return nil
			}

			success, err := replayWorkflowExecutionHelper(
				ctx,
				s.replayer,
				s.service,
				s.logger,
				s.domain,
				WorkflowExecution{
					ID:    execution.GetWorkflowId(),
					RunID: execution.GetRunId(),
				},
			)
			if err != nil {
				return err
			}
			if success {
				replayCount++
			}

			if replayCount == maxReplayCount {
				return nil
			}
		}

		if len(scanResult.NextPageToken) == 0 {
			if s.shadowOptions.Mode == ShadowModeNormal || s.clock.Now().Add(defaultWaitDurationPerIteration).After(expirationTime) {
				return nil
			}

			s.clock.Sleep(defaultWaitDurationPerIteration)
		}

		scanRequest.NextPageToken = scanResult.NextPageToken
	}

}

func (o *ShadowOptions) validateAndPopulateFields() error {
	exitConditionSpecified := o.ExitCondition.ExpirationInterval > 0 || o.ExitCondition.ShadowCount > 0
	if o.Mode == ShadowModeContinuous && !exitConditionSpecified {
		return errors.New("exit condition must be specified if shadow mode is set to continuous")
	}

	if o.SamplingRate < 0 || o.SamplingRate > 1 {
		return errors.New("sampling rate should be in range [0, 1]")
	}

	if len(o.WorkflowQuery) != 0 && (len(o.WorkflowTypes) != 0 || len(o.WorkflowStatus) != 0 || !o.WorkflowStartTimeFilter.isEmpty()) {
		return errors.New("workflow types, status, start time and close time filter can't be specified when workflow query is specified")
	}

	if len(o.WorkflowQuery) == 0 {
		queryBuilder := NewQueryBuilder()

		if len(o.WorkflowTypes) > 0 {
			queryBuilder.WorkflowTypes(o.WorkflowTypes)
		}

		if len(o.ExcludeTypes) > 0 {
			queryBuilder.ExcludeWorkflowTypes(o.ExcludeTypes)
		}

		statuses := make([]WorkflowStatus, 0, len(o.WorkflowStatus))
		for _, statusString := range o.WorkflowStatus {
			status, err := ToWorkflowStatus(statusString)
			if err != nil {
				return err
			}
			statuses = append(statuses, status)
		}
		// TODO: handle ContinueAsNew as a special case.
		// This filter doesn't seem to be working as expected.
		// This filter should have only taken open and closed wfs but it returns continue as new wfs as well.
		if len(statuses) == 0 {
			statuses = []WorkflowStatus{WorkflowStatusOpen, WorkflowStatusClosed}
		}

		queryBuilder.WorkflowStatus(statuses)

		if !o.WorkflowStartTimeFilter.isEmpty() {
			if err := o.WorkflowStartTimeFilter.validateAndPopulateFields(); err != nil {
				return fmt.Errorf("invalid start time filter, error: %v", err)
			}
			queryBuilder.StartTime(o.WorkflowStartTimeFilter.MinTimestamp, o.WorkflowStartTimeFilter.MaxTimestamp)
		}

		if !o.WorkflowCloseTimeFilter.isEmpty() {
			if err := o.WorkflowCloseTimeFilter.validateAndPopulateFields(); err != nil {
				return fmt.Errorf("invalid close time filter, error: %v", err)
			}
			queryBuilder.CloseTime(o.WorkflowCloseTimeFilter.MinTimestamp, o.WorkflowCloseTimeFilter.MaxTimestamp)
		}
		o.WorkflowQuery = queryBuilder.Build()
	}

	if o.SamplingRate == 0 {
		// if not set, defaults to replay all workflows
		o.SamplingRate = 1
	}

	if o.Concurrency == 0 {
		// if not set, defaults to 1
		o.Concurrency = 1
	}

	return nil
}

func (t *TimeFilter) validateAndPopulateFields() error {
	if t.MaxTimestamp.IsZero() {
		t.MaxTimestamp = maxTimestamp
	}

	if t.MaxTimestamp.Before(t.MinTimestamp) {
		return errors.New("maxTimestamp should be after minTimestamp")
	}

	return nil
}

func (t *TimeFilter) isEmpty() bool {
	if t == nil {
		return true
	}

	return t.MinTimestamp.IsZero() && t.MaxTimestamp.IsZero()
}

func (m ShadowMode) toThriftPtr() *shadower.Mode {
	switch m {
	case ShadowModeNormal:
		return shadower.ModeNormal.Ptr()
	case ShadowModeContinuous:
		return shadower.ModeContinuous.Ptr()
	default:
		panic(fmt.Sprintf("unknown shadow mode %v", m))
	}
}

func (e ShadowExitCondition) toThriftPtr() *shadower.ExitCondition {
	return &shadower.ExitCondition{
		ShadowCount:                 common.Int32Ptr(int32(e.ShadowCount)),
		ExpirationIntervalInSeconds: common.Int32Ptr(int32(e.ExpirationInterval.Seconds())),
	}
}
