// Copyright (c) 2017 Uber Technologies, Inc.
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

// when adding any, make sure you update the files that it checks in the makefile
//go:generate mockery --srcpkg . --name Client --output ../mocks
//go:generate mockery --srcpkg . --name DomainClient --output ../mocks

// Package client contains functions to create Cadence clients used to communicate to Cadence service.
//
// Use these to perform CRUD on domains and start or query workflow executions.
package client

import (
	"context"
	"errors"

	"go.uber.org/cadence"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/internal"
	"go.uber.org/cadence/workflow"
)

const (
	// QueryTypeStackTrace is the build in query type for Client.QueryWorkflow() call. Use this query type to get the call
	// stack of the workflow. The result will be a string encoded in the encoded.Value.
	QueryTypeStackTrace string = internal.QueryTypeStackTrace

	// QueryTypeOpenSessions is the build in query type for Client.QueryWorkflow() call. Use this query type to get all open
	// sessions in the workflow. The result will be a list of SessionInfo encoded in the encoded.Value.
	QueryTypeOpenSessions string = internal.QueryTypeOpenSessions

	// QueryTypeQueryTypes is the build in query type for Client.QueryWorkflow() call. Use this query type to list
	// all query types of the workflow. The result will be a string encoded in the EncodedValue.
	QueryTypeQueryTypes string = internal.QueryTypeQueryTypes
)

type (
	// Options are optional parameters for Client creation.
	Options = internal.ClientOptions

	// FeatureFlags define which breaking changes can be enabled for client
	FeatureFlags = internal.FeatureFlags

	// StartWorkflowOptions configuration parameters for starting a workflow execution.
	StartWorkflowOptions = internal.StartWorkflowOptions

	// HistoryEventIterator is a iterator which can return history events
	HistoryEventIterator = internal.HistoryEventIterator

	// WorkflowRun represents a started non child workflow
	WorkflowRun = internal.WorkflowRun

	// WorkflowIDReusePolicy defines workflow ID reuse behavior.
	WorkflowIDReusePolicy = internal.WorkflowIDReusePolicy

	// QueryWorkflowWithOptionsRequest defines the request to QueryWorkflowWithOptions
	QueryWorkflowWithOptionsRequest = internal.QueryWorkflowWithOptionsRequest

	// QueryWorkflowWithOptionsResponse defines the response to QueryWorkflowWithOptions
	QueryWorkflowWithOptionsResponse = internal.QueryWorkflowWithOptionsResponse

	// GetWorkflowHistoryWithOptionsRequest defines the request to GetWorkflowHistoryWithOptions
	GetWorkflowHistoryWithOptionsRequest = internal.GetWorkflowHistoryWithOptionsRequest

	// DescribeWorkflowExecutionWithOptionsRequest defines the request to DescribeWorkflowExecutionWithOptions
	DescribeWorkflowExecutionWithOptionsRequest = internal.DescribeWorkflowExecutionWithOptionsRequest

	// ParentClosePolicy defines the behavior performed on a child workflow when its parent is closed
	ParentClosePolicy = internal.ParentClosePolicy

	// QueryConsistencyLevel defines the level of consistency the query should respond with.
	// It will default to the cluster's configuration if not specified.
	// Valid values are QueryConsistencyLevelEventual (served by the receiving cluster), and QueryConsistencyLevelStrong (redirects to the active cluster).
	QueryConsistencyLevel = internal.QueryConsistencyLevel

	// CancelOption values are functional options for the CancelWorkflow method.
	// Supported values can be created with:
	//  - WithCancelReason(...)
	CancelOption = internal.Option

	// Client is the client for starting and getting information about a workflow executions as well as
	// completing activities asynchronously.
	Client interface {
		// StartWorkflow starts a workflow execution
		// The user can use this to start using a function or workflow type name.
		// Either by
		//     StartWorkflow(ctx, options, "workflowTypeName", arg1, arg2, arg3)
		//     or
		//     StartWorkflow(ctx, options, workflowExecuteFn, arg1, arg2, arg3)
		// The errors it can return:
		//	- EntityNotExistsError, if domain does not exists
		//	- BadRequestError
		//	- WorkflowExecutionAlreadyStartedError
		//	- InternalServiceError
		StartWorkflow(ctx context.Context, options StartWorkflowOptions, workflowFunc interface{}, args ...interface{}) (*workflow.Execution, error)

		// StartWorkflowAsync behaves like StartWorkflow except that the request is first queued and then processed asynchronously.
		// See StartWorkflow for parameter details.
		// The returned AsyncWorkflowExecution doesn't contain run ID, because the workflow hasn't started yet.
		// The errors it can return:
		//	- EntityNotExistsError, if domain does not exists
		//	- BadRequestError
		//	- InternalServiceError
		StartWorkflowAsync(ctx context.Context, options StartWorkflowOptions, workflow interface{}, args ...interface{}) (*workflow.ExecutionAsync, error)

		// ExecuteWorkflow starts a workflow execution and return a WorkflowRun instance and error
		// The user can use this to start using a function or workflow type name.
		// Either by
		//     ExecuteWorkflow(ctx, options, "workflowTypeName", arg1, arg2, arg3)
		//     or
		//     ExecuteWorkflow(ctx, options, workflowExecuteFn, arg1, arg2, arg3)
		// The errors it can return:
		//	- EntityNotExistsError, if domain does not exists
		//	- BadRequestError
		//	- InternalServiceError
		//
		// WorkflowRun has 2 methods:
		//  - GetRunID() string: which return the first started workflow run ID (please see below)
		//  - Get(ctx context.Context, valuePtr interface{}) error: which will fill the workflow
		//    execution result to valuePtr, if workflow execution is a success, or return corresponding
		//    error. This is a blocking API.
		// NOTE: if the started workflow return ContinueAsNewError during the workflow execution, the
		// return result of GetRunID() will be the started workflow run ID, not the new run ID caused by ContinueAsNewError,
		// however, Get(ctx context.Context, valuePtr interface{}) will return result from the run which did not return ContinueAsNewError.
		// Say ExecuteWorkflow started a workflow, in its first run, has run ID "run ID 1", and returned ContinueAsNewError,
		// the second run has run ID "run ID 2" and return some result other than ContinueAsNewError:
		// GetRunID() will always return "run ID 1" and  Get(ctx context.Context, valuePtr interface{}) will return the result of second run.
		// NOTE: DO NOT USE THIS API INSIDE A WORKFLOW, USE workflow.ExecuteChildWorkflow instead
		ExecuteWorkflow(ctx context.Context, options StartWorkflowOptions, workflow interface{}, args ...interface{}) (WorkflowRun, error)

		// GetWorkflow retrieves a workflow execution and return a WorkflowRun instance (described above)
		// - workflow ID of the workflow.
		// - runID can be default(empty string). if empty string then it will pick the last running execution of that workflow ID.
		//
		// WorkflowRun has 2 methods:
		//  - GetRunID() string: which return the first started workflow run ID (please see below)
		//  - Get(ctx context.Context, valuePtr interface{}) error: which will fill the workflow
		//    execution result to valuePtr, if workflow execution is a success, or return corresponding
		//    error. This is a blocking API.
		// If workflow not found, the Get() will return EntityNotExistsError.
		// NOTE: if the started workflow return ContinueAsNewError during the workflow execution, the
		// return result of GetRunID() will be the started workflow run ID, not the new run ID caused by ContinueAsNewError,
		// however, Get(ctx context.Context, valuePtr interface{}) will return result from the run which did not return ContinueAsNewError.
		// Say ExecuteWorkflow started a workflow, in its first run, has run ID "run ID 1", and returned ContinueAsNewError,
		// the second run has run ID "run ID 2" and return some result other than ContinueAsNewError:
		// GetRunID() will always return "run ID 1" and  Get(ctx context.Context, valuePtr interface{}) will return the result of second run.
		GetWorkflow(ctx context.Context, workflowID string, runID string) WorkflowRun

		// SignalWorkflow sends a signals to a workflow in execution
		// - workflow ID of the workflow.
		// - runID can be default(empty string). if empty string then it will pick the running execution of that workflow ID.
		// - signalName name to identify the signal.
		// - arg is the data (or nil) to send with the signal, which can be read with the signal channel's Receive out-arg.
		// The errors it can return:
		//	- EntityNotExistsError
		//	- InternalServiceError
		//	- WorkflowExecutionAlreadyCompletedError
		SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error

		// SignalWithStartWorkflow sends a signal to a running workflow.
		// If the workflow is not running or not found, it starts the workflow and then sends the signal in transaction.
		// - workflowID, signalName, signalArg are same as SignalWorkflow's parameters
		// - options, workflow, workflowArgs are same as StartWorkflow's parameters
		// The errors it can return:
		//  - EntityNotExistsError, if domain does not exist
		//  - BadRequestError
		//	- InternalServiceError
		SignalWithStartWorkflow(ctx context.Context, workflowID string, signalName string, signalArg interface{},
			options StartWorkflowOptions, workflowFunc interface{}, workflowArgs ...interface{}) (*workflow.Execution, error)

		// SignalWithStartWorkflowAsync behaves like SignalWithStartWorkflow except that the request is first queued and then processed asynchronously.
		// See SignalWithStartWorkflow for parameter details.
		// The errors it can return:
		//  - EntityNotExistsError, if domain does not exist
		//  - BadRequestError
		//	- InternalServiceError
		SignalWithStartWorkflowAsync(ctx context.Context, workflowID string, signalName string, signalArg interface{},
			options StartWorkflowOptions, workflow interface{}, workflowArgs ...interface{}) (*workflow.ExecutionAsync, error)

		// CancelWorkflow cancels a workflow in execution
		// - workflow ID of the workflow.
		// - runID can be default(empty string). if empty string then it will pick the running execution of that workflow ID.
		// The errors it can return:
		//	- EntityNotExistsError
		//	- BadRequestError
		//	- InternalServiceError
		//	- WorkflowExecutionAlreadyCompletedError
		CancelWorkflow(ctx context.Context, workflowID string, runID string, opts ...CancelOption) error

		// TerminateWorkflow terminates a workflow execution.
		// workflowID is required, other parameters are optional.
		// - workflow ID of the workflow.
		// - runID can be default(empty string). if empty string then it will pick the running execution of that workflow ID.
		// The errors it can return:
		//	- EntityNotExistsError
		//	- BadRequestError
		//	- InternalServiceError
		//	- WorkflowExecutionAlreadyCompletedError
		TerminateWorkflow(ctx context.Context, workflowID string, runID string, reason string, details []byte) error

		// GetWorkflowHistory gets history events of a particular workflow
		// - workflow ID of the workflow.
		// - runID can be default(empty string). if empty string then it will pick the last running execution of that workflow ID.
		// - whether use long poll for tracking new events: when the workflow is running, there can be new events generated during iteration
		// 	 of HistoryEventIterator, if isLongPoll == true, then iterator will do long poll, tracking new history event, i.e. the iteration
		//   will not be finished until workflow is finished; if isLongPoll == false, then iterator will only return current history events.
		// - whether return all history events or just the last event, which contains the workflow execution end result
		// Example:-
		//	To iterate all events,
		// 		iter := GetWorkflowHistory(ctx, workflowID, runID, isLongPoll, filterType)
		//		events := []*shared.HistoryEvent{}
		//		for iter.HasNext() {
		//			event, err := iter.Next()
		//			if err != nil {
		//				return err
		//			}
		//			events = append(events, event)
		//		}
		GetWorkflowHistory(ctx context.Context, workflowID string, runID string, isLongPoll bool, filterType s.HistoryEventFilterType) HistoryEventIterator

		// GetWorkflowHistoryWithOptions gets history events of a particular workflow.
		// See GetWorkflowHistoryWithOptionsRequest for more information.
		// Returns an iterator of HistoryEvents - see shared.HistoryEvent for more details.
		// Example:-
		//	To iterate all events,
		// 		iter := GetWorkflowHistory(ctx, workflowID, runID, isLongPoll, filterType)
		//		events := []*shared.HistoryEvent{}
		//		for iter.HasNext() {
		//			event, err := iter.Next()
		//			if err != nil {
		//				return err
		//			}
		//			events = append(events, event)
		//		}
		// Returns the following errors:
		//  - EntityNotExistsError
		//  - BadRequestError
		//  - InternalServiceError
		GetWorkflowHistoryWithOptions(ctx context.Context, request *GetWorkflowHistoryWithOptionsRequest) HistoryEventIterator

		// CompleteActivity reports activity completed.
		// activity Execute method can return activity.ErrResultPending to
		// indicate the activity is not completed when it's Execute method returns. In that case, this CompleteActivity() method
		// should be called when that activity is completed with the actual result and error. If err is nil, activity task
		// completed event will be reported; if err is CanceledError, activity task cancelled event will be reported; otherwise,
		// activity task failed event will be reported.
		// An activity implementation should use GetActivityInfo(ctx).TaskToken function to get task token to use for completion.
		// Example:-
		//	To complete with a result.
		//  	CompleteActivity(token, "Done", nil)
		//	To fail the activity with an error.
		//      CompleteActivity(token, nil, cadence.NewCustomError("reason", details)
		// The activity can fail with below errors ErrorWithDetails, TimeoutError, CanceledError.
		CompleteActivity(ctx context.Context, taskToken []byte, result interface{}, err error) error

		// CompleteActivityByID reports activity completed.
		// Similar to CompleteActivity, but may save cadence user from keeping taskToken info.
		// activity Execute method can return activity.ErrResultPending to
		// indicate the activity is not completed when it's Execute method returns. In that case, this CompleteActivityById() method
		// should be called when that activity is completed with the actual result and error. If err is nil, activity task
		// completed event will be reported; if err is CanceledError, activity task cancelled event will be reported; otherwise,
		// activity task failed event will be reported.
		// An activity implementation should use activityID provided in ActivityOption to use for completion.
		// domain name, workflowID, activityID are required, runID is optional.
		// The errors it can return:
		//  - ErrorWithDetails
		//  - TimeoutError
		//  - CanceledError
		CompleteActivityByID(ctx context.Context, domain, workflowID, runID, activityID string, result interface{}, err error) error

		// RecordActivityHeartbeat records heartbeat for an activity.
		// taskToken - is the value of the binary "TaskToken" field of the "ActivityInfo" struct retrieved inside the activity.
		// details - is the progress you want to record along with heart beat for this activity.
		// The errors it can return:
		//	- EntityNotExistsError
		//	- InternalServiceError
		RecordActivityHeartbeat(ctx context.Context, taskToken []byte, details ...interface{}) error

		// RecordActivityHeartbeatByID records heartbeat for an activity.
		// details - is the progress you want to record along with heart beat for this activity.
		// The errors it can return:
		//	- EntityNotExistsError
		//	- InternalServiceError
		RecordActivityHeartbeatByID(ctx context.Context, domain, workflowID, runID, activityID string, details ...interface{}) error

		// ListClosedWorkflow gets closed workflow executions based on request filters.
		// Retrieved workflow executions are sorted by start time in descending order.
		// (Retrieved workflow executions could also be sorted by closed time in descending order,
		// if cadence server side config EnableReadFromClosedExecutionV2 is set to true.)
		// Note: heavy usage of this API may cause huge persistence pressure.
		// The errors it can return:
		//  - BadRequestError
		//  - InternalServiceError
		//  - EntityNotExistError
		ListClosedWorkflow(ctx context.Context, request *s.ListClosedWorkflowExecutionsRequest) (*s.ListClosedWorkflowExecutionsResponse, error)

		// ListOpenWorkflow gets open workflow executions based on request filters.
		// Retrieved workflow executions are sorted by start time in descending order.
		// Note: heavy usage of this API may cause huge persistence pressure.
		// The errors it can return:
		//  - BadRequestError
		//  - InternalServiceError
		//  - EntityNotExistError
		ListOpenWorkflow(ctx context.Context, request *s.ListOpenWorkflowExecutionsRequest) (*s.ListOpenWorkflowExecutionsResponse, error)

		// ListWorkflow gets workflow executions based on query. This API only works with ElasticSearch,
		// and will return BadRequestError when using Cassandra or MySQL. The query is basically the SQL WHERE clause,
		// examples:
		//  - "(WorkflowID = 'wid1' or (WorkflowType = 'type2' and WorkflowID = 'wid2'))".
		//  - "CloseTime between '2019-08-27T15:04:05+00:00' and '2019-08-28T15:04:05+00:00'".
		//  - to list only open workflow use "CloseTime = missing"
		// Retrieved workflow executions are sorted by StartTime in descending order when list open workflow,
		// and sorted by CloseTime in descending order for other queries.
		// The errors it can return:
		//  - BadRequestError
		//  - InternalServiceError
		ListWorkflow(ctx context.Context, request *s.ListWorkflowExecutionsRequest) (*s.ListWorkflowExecutionsResponse, error)

		// ListArchivedWorkflow gets archived workflow executions based on query. This API will return BadRequest if Cadence
		// cluster or target domain is not configured for visibility archival or read is not enabled. The query is basically the SQL WHERE clause.
		// However, different visibility archivers have different limitations on the query. Please check the documentation of the visibility archiver used
		// by your domain to see what kind of queries are accept and whether retrieved workflow executions are ordered or not.
		// The errors it can return:
		//  - BadRequestError
		//  - InternalServiceError
		ListArchivedWorkflow(ctx context.Context, request *s.ListArchivedWorkflowExecutionsRequest) (*s.ListArchivedWorkflowExecutionsResponse, error)

		// ScanWorkflow gets workflow executions based on query. This API only works with ElasticSearch,
		// and will return BadRequestError when using Cassandra or MySQL. The query is basically the SQL WHERE clause
		// (see ListWorkflow for query examples).
		// ScanWorkflow should be used when retrieving large amount of workflows and order is not needed.
		// It will use more ElasticSearch resources than ListWorkflow, but will be several times faster
		// when retrieving millions of workflows.
		// The errors it can return:
		//  - BadRequestError
		//  - InternalServiceError
		ScanWorkflow(ctx context.Context, request *s.ListWorkflowExecutionsRequest) (*s.ListWorkflowExecutionsResponse, error)

		// CountWorkflow gets number of workflow executions based on query. This API only works with ElasticSearch,
		// and will return BadRequestError when using Cassandra or MySQL. The query is basically the SQL WHERE clause
		// (see ListWorkflow for query examples).
		// The errors it can return:
		//  - BadRequestError
		//  - InternalServiceError
		CountWorkflow(ctx context.Context, request *s.CountWorkflowExecutionsRequest) (*s.CountWorkflowExecutionsResponse, error)

		// GetSearchAttributes returns valid search attributes keys and value types.
		// The search attributes can be used in query of List/Scan/Count APIs. Adding new search attributes requires cadence server
		// to update dynamic config ValidSearchAttributes.
		GetSearchAttributes(ctx context.Context) (*s.GetSearchAttributesResponse, error)

		// QueryWorkflow queries a given workflow's last execution and returns the query result synchronously. Parameter workflowID
		// and queryType are required, other parameters are optional. The workflowID and runID (optional) identify the
		// target workflow execution that this query will be send to. If runID is not specified (empty string), server will
		// use the currently running execution of that workflowID. The queryType specifies the type of query you want to
		// run. By default, cadence supports "__stack_trace" as a standard query type, which will return string value
		// representing the call stack of the target workflow. The target workflow could also setup different query handler
		// to handle custom query types.
		// See comments at workflow.SetQueryHandler(ctx Context, queryType string, handler interface{}) for more details
		// on how to setup query handler within the target workflow.
		// - workflowID is required.
		// - runID can be default(empty string). if empty string then it will pick the running execution of that workflow ID.
		// - queryType is the type of the query.
		// - args... are the optional query parameters, which will be sent to your query handler function's args.
		// The errors it can return:
		//  - BadRequestError
		//  - InternalServiceError
		//  - EntityNotExistError
		//  - QueryFailError
		QueryWorkflow(ctx context.Context, workflowID string, runID string, queryType string, args ...interface{}) (encoded.Value, error)

		// QueryWorkflowWithOptions queries a given workflow execution and returns the query result synchronously.
		// See QueryWorkflowWithOptionsRequest and QueryWorkflowWithOptionsResponse for more information.
		// The errors it can return:
		//  - BadRequestError
		//  - InternalServiceError
		//  - EntityNotExistError
		//  - QueryFailError
		QueryWorkflowWithOptions(ctx context.Context, request *QueryWorkflowWithOptionsRequest) (*QueryWorkflowWithOptionsResponse, error)

		// ResetWorkflow reset a given workflow execution and returns a new execution
		// See ResetWorkflowRequest and ResetWorkflowResponse for more information.
		// The errors it can return:
		//  - BadRequestError
		//  - InternalServiceError
		//  - EntityNotExistError
		ResetWorkflow(ctx context.Context, request *s.ResetWorkflowExecutionRequest) (*s.ResetWorkflowExecutionResponse, error)

		// DescribeWorkflowExecution returns information about the specified workflow execution.
		// - runID can be default(empty string). if empty string then it will pick the last running execution of that workflow ID.
		//
		// The errors it can return:
		//  - BadRequestError
		//  - InternalServiceError
		//  - EntityNotExistError
		DescribeWorkflowExecution(ctx context.Context, workflowID, runID string) (*s.DescribeWorkflowExecutionResponse, error)

		// DescribeWorkflowExecutionWithOptions returns information about the specified workflow execution.
		// See DescribeWorkflowExecutionWithOptionsRequest for more information.
		// The errors it can return:
		//  - BadRequestError
		//  - InternalServiceError
		//  - EntityNotExistError
		DescribeWorkflowExecutionWithOptions(ctx context.Context, request *DescribeWorkflowExecutionWithOptionsRequest) (*s.DescribeWorkflowExecutionResponse, error)

		// DescribeTaskList returns information about the target tasklist, right now this API returns the
		// pollers which polled this tasklist in last few minutes.
		// The errors it can return:
		//  - BadRequestError
		//  - InternalServiceError
		//  - EntityNotExistError
		DescribeTaskList(ctx context.Context, tasklist string, tasklistType s.TaskListType) (*s.DescribeTaskListResponse, error)

		// RefreshWorkflowTasks refreshes all the tasks of a given workflow.
		// - workflow ID of the workflow.
		// - runID can be default(empty string). if empty string then it will pick the running execution of that workflow ID.
		// The errors it can return:
		//  - BadRequestError
		//  - DomainNotActiveError
		//  - ServiceBusyError
		//  - EntityNotExistError
		RefreshWorkflowTasks(ctx context.Context, workflowID, runID string) error
	}

	// DomainClient is the client for managing operations on the domain.
	// CLI, tools, ... can use this layer to manager operations on domain.
	DomainClient interface {
		// Register a domain with cadence server
		// The errors it can throw:
		//	- DomainAlreadyExistsError
		//	- BadRequestError
		//	- InternalServiceError
		Register(ctx context.Context, request *s.RegisterDomainRequest) error

		// Describe a domain. The domain has 3 part of information
		// DomainInfo - Which has Name, Status, Description, Owner Email
		// DomainConfiguration - Configuration like Workflow Execution Retention Period In Days, Whether to emit metrics.
		// ReplicationConfiguration - replication config like clusters and active cluster name
		// The errors it can throw:
		//	- EntityNotExistsError
		//	- BadRequestError
		//	- InternalServiceError
		Describe(ctx context.Context, name string) (*s.DescribeDomainResponse, error)

		// Update a domain.
		// The errors it can throw:
		//	- EntityNotExistsError
		//	- BadRequestError
		//	- InternalServiceError
		Update(ctx context.Context, request *s.UpdateDomainRequest) error
	}
)

const (
	// WorkflowIDReusePolicyAllowDuplicateFailedOnly allow start a workflow execution
	// when workflow not running, and the last execution close state is in
	// [terminated, cancelled, timeout, failed].
	WorkflowIDReusePolicyAllowDuplicateFailedOnly WorkflowIDReusePolicy = internal.WorkflowIDReusePolicyAllowDuplicateFailedOnly

	// WorkflowIDReusePolicyAllowDuplicate allow start a workflow execution using
	// the same workflow ID,when workflow not running.
	WorkflowIDReusePolicyAllowDuplicate WorkflowIDReusePolicy = internal.WorkflowIDReusePolicyAllowDuplicate

	// WorkflowIDReusePolicyRejectDuplicate do not allow start a workflow execution using the same workflow ID at all
	WorkflowIDReusePolicyRejectDuplicate WorkflowIDReusePolicy = internal.WorkflowIDReusePolicyRejectDuplicate

	// WorkflowIDReusePolicyTerminateIfRunning terminate current running workflow using the same workflow ID if exist,
	// then start a new run in one transaction
	WorkflowIDReusePolicyTerminateIfRunning = internal.WorkflowIDReusePolicyTerminateIfRunning
)

const (
	// ParentClosePolicyTerminate means terminating the child workflow
	ParentClosePolicyTerminate = internal.ParentClosePolicyTerminate
	// ParentClosePolicyRequestCancel means requesting cancellation on the child workflow
	ParentClosePolicyRequestCancel = internal.ParentClosePolicyRequestCancel
	// ParentClosePolicyAbandon means not doing anything on the child workflow
	ParentClosePolicyAbandon = internal.ParentClosePolicyAbandon
)

const (
	// QueryConsistencyLevelUnspecified will use the default consistency level provided by the cluster.
	QueryConsistencyLevelUnspecified = internal.QueryConsistencyLevelUnspecified
	// QueryConsistencyLevelEventual passes the request to the receiving cluster and returns eventually consistent results
	QueryConsistencyLevelEventual = internal.QueryConsistencyLevelEventual
	// QueryConsistencyLevelStrong will redirect the request to the active cluster and returns strongly consistent results
	QueryConsistencyLevelStrong = internal.QueryConsistencyLevelStrong
)

// NewClient creates an instance of a workflow client
func NewClient(service workflowserviceclient.Interface, domain string, options *Options) Client {
	return internal.NewClient(service, domain, options)
}

// NewDomainClient creates an instance of a domain client, to manage lifecycle of domains.
func NewDomainClient(service workflowserviceclient.Interface, options *Options) DomainClient {
	return internal.NewDomainClient(service, options)
}

// make sure if new methods are added to internal.Client they are also added to public Client.
var _ Client = internal.Client(nil)
var _ internal.Client = Client(nil)
var _ DomainClient = internal.DomainClient(nil)
var _ internal.DomainClient = DomainClient(nil)

// NewValue creates a new encoded.Value which can be used to decode binary data returned by Cadence.  For example:
// User had Activity.RecordHeartbeat(ctx, "my-heartbeat") and then got response from calling Client.DescribeWorkflowExecution.
// The response contains binary field PendingActivityInfo.HeartbeatDetails,
// which can be decoded by using:
//
//	var result string // This need to be same type as the one passed to RecordHeartbeat
//	NewValue(data).Get(&result)
func NewValue(data []byte) encoded.Value {
	return internal.NewValue(data)
}

// NewValues creates a new encoded.Values which can be used to decode binary data returned by Cadence. For example:
// User had Activity.RecordHeartbeat(ctx, "my-heartbeat", 123) and then got response from calling Client.DescribeWorkflowExecution.
// The response contains binary field PendingActivityInfo.HeartbeatDetails,
// which can be decoded by using:
//
//	var result1 string
//	var result2 int // These need to be same type as those arguments passed to RecordHeartbeat
//	NewValues(data).Get(&result1, &result2)
func NewValues(data []byte) encoded.Values {
	return internal.NewValues(data)
}

// IsWorkflowError returns true if an error is a known returned-by-the-workflow error type, when it comes from
// WorkflowRun from either Client.ExecuteWorkflow or Client.GetWorkflow.  If it returns false, the error comes from
// some other source, e.g. RPC request failures or bad arguments.
// Using it on errors from any other source is currently undefined.
//
// This checks for known types via errors.As, so it will check recursively if it encounters wrapped errors.
// Note that this is different than the various `cadence.Is[Type]Error` checks, which do not unwrap.
//
// Currently the complete list of errors checked is:
// *cadence.CustomError, *cadence.CanceledError, *workflow.ContinueAsNewError,
// *workflow.GenericError, *workflow.TimeoutError, *workflow.TerminatedError,
// *workflow.PanicError, *workflow.UnknownExternalWorkflowExecutionError
//
// See documentation for each error type for details.
func IsWorkflowError(err error) bool {
	var custom *cadence.CustomError
	if errors.As(err, &custom) {
		return true
	}
	var cancel *cadence.CanceledError
	if errors.As(err, &cancel) {
		return true
	}

	var generic *workflow.GenericError
	if errors.As(err, &generic) {
		return true
	}
	var timeout *workflow.TimeoutError
	if errors.As(err, &timeout) {
		return true
	}
	var terminate *workflow.TerminatedError
	if errors.As(err, &terminate) {
		return true
	}
	var panicked *workflow.PanicError
	if errors.As(err, &panicked) {
		return true
	}
	var can *workflow.ContinueAsNewError
	if errors.As(err, &can) {
		return true
	}
	var unknown *workflow.UnknownExternalWorkflowExecutionError
	if errors.As(err, &unknown) {
		return true
	}
	return false
}

// WithCancelReason can be passed to Client.CancelWorkflow to provide an explicit cancellation reason,
// which will be recorded in the cancellation event in the workflow's history, similar to termination reasons.
// This is purely informational, and does not influence Cadence behavior at all.
func WithCancelReason(reason string) CancelOption {
	return internal.WithCancelReason(reason)
}
