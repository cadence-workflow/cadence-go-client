// Copyright (c) 2021 Uber Technologies, Inc.
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

package client

import (
	"context"
	"time"

	"go.uber.org/cadence/encoded"
)

type (
	// Client is the top level client entity for interactive with Cadence server.
	Client interface {
		// Domains returns an interface for interacting with Cadence domains.
		Domains() Domains

		// Activities returns an interface for interacting with activities (via opaque task tokens).
		Activities() Activities

		// SearchAttributes returns an interface for interacting with cluster search attributes.
		SearchAttributes() SearchAttributes

		// QueryBuilder returns a query builder to contruct workflow queries
		QueryBuilder() QueryBuilder
	}

	// Domains is an interface for interacting with Cadence domains.
	Domains interface {
		// Register registers a domain within cadence server.
		// Name and replication settings are required. Optional fields can be set via options.
		Register(ctx context.Context, name string, replication DomainReplicationConfig, opts ...DomainRegisterOption) (Domain, error)

		// List retrieves all domains that are registered within Cadence server.
		List(ctx context.Context, page Page, opts ...DomainListOption) ([]ListedDomain, Page, error)

		// Get selects a domain by a given name for further operations.
		Get(name string) Domain
	}

	// ListedDomain is a domain that was retrieved via List call and contains additional info.
	ListedDomain interface {
		Domain
		Info() DomainInfo
	}

	// Domain is a Cadence way to group workflows of some application or owner.
	Domain interface {
		// Name return the name of this domain
		Name() string

		// Describe returns information about the domain.
		Describe(ctx context.Context, opts ...DomainDescribeOption) (DomainInfo, error)

		// Update updates one or more domain fields. Use options to specify which fields to update.
		Update(ctx context.Context, opts ...DomainUpdateOption) error

		// Failover will failover the domain to another given cluster.
		Failover(ctx context.Context, cluster string, opts ...DomainFailoverOption) error

		// Deprecate will deprecate the domain.
		Deprecate(ctx context.Context, opts ...DomainDeprecateOption) error

		// AddBadBinary will include new bad binary with the given checksum.
		AddBadBinary(ctx context.Context, checksum string, reason string, opts ...DomainAddBadBinaryOption) error

		// DeleteBadBinary will delete an existing bad binary by the given checksum.
		DeleteBadBinary(ctx context.Context, checksum string, opts ...DomainDeleteBadBinaryOption) error
	}

	// Workflows is an interface for interacting with Cadence workflows for the selected domain.
	Workflows interface {
		// Starts starts a new workflow execution
		// The user can use this to start using a function or workflow type name.
		Start(ctx context.Context, wfFunc interface{}, args []interface{}, taskList string, timeout time.Duration, opts ...WorkflowStartOption) (Workflow, error)

		// Count gets a number of workflow executions based on query.
		Count(ctx context.Context, query Query, opts ...WorkflowCountOption) (int64, error)

		// List returns a list of workflow executions based on query.
		List(ctx context.Context, query Query, page Page, opts ...WorkflowListOption) ([]ListedWorkflow, *Page, error)

		// Get will select a concrete workflow run by a given workflow and run ID for further operations.
		Get(workflowID, runID string) Workflow

		// GetCurrent will select the latest run of the workflow given by its workflowID for further operations.
		GetCurrent(workflowID string) Workflow
	}

	// ListedWorkflow is a workflow that was retrieved via List call and contains additional info.
	ListedWorkflow interface {
		Workflow
		Info() WorkflowInfo
	}

	// Workflow
	Workflow interface {
		// Domain returns Domain of this workflow
		Domain() Domain
		// WorkflowID returns ID of this workflow
		WorkflowID() string
		// RunID return Run ID of this workflow
		RunID() string

		// Signal sends a signals to a workflow in execution.
		Signal(ctx context.Context, signalName string, args []interface{}, opts ...WorkflowSignalOption) error

		// Query queries a workflow's execution and returns the query result synchronously.
		// The queryType specifies the type of query you want to run.
		// By default, cadence supports "__stack_trace" as a standard query type, which will return string value
		// representing the call stack of the target workflow. The target workflow could also setup different query handler
		// to handle custom query types.
		// See comments at workflow.SetQueryHandler(ctx Context, queryType string, handler interface{}) for more details
		// on how to setup query handler within the target workflow.
		Query(ctx context.Context, queryType string, args []interface{}, opts ...WorkflowQueryOption) (encoded.Value, error)

		// Describe returns information about the workflow execution.
		Describe(ctx context.Context, opts ...WorkflowDescribeOption) (ExtendedWorkflowInfo, error)

		// Cancel cancels a workflow in execution.
		// This is different from Terminate, as it would cancel workflow context giving opportunity for it to do the cleanup and exit gracefully.
		Cancel(ctx context.Context, opts ...WorkflowCancelOption) error

		// Terminate terminates a workflow execution.
		// This is different from Cancel, as it would force workflow termination from server side.
		Terminate(ctx context.Context, reason string, opts ...WorkflowTerminateOption) error

		// Reset resets a workflow execution to the given point and returns a new execution.
		Reset(ctx context.Context, reason string, point WorkflowResetPoint, opts ...WorkflowResetOption) (Workflow, error)

		// GetResult will return workflow result if workflow execution is a success,
		// or return corresponding error. This is a blocking API.
		GetResult(ctx context.Context, opts ...WorkflowGetResultOption) (encoded.Value, error)

		// Activities returns an interface for interactive activities of this workflow.
		Activities() WorkflowActivities
	}

	// Activities is an interface for interacting with any activities.
	Activities interface {
		// Get selects an activity with the given task token.
		// This is an opaque token that can be obtained within activity via GetActivityInfo(ctx).TaskToken function.
		// It is used to interact with activity externally - to complete it or record is progress and heartbeat.
		Get(taskToken []byte) Activity
	}

	// WorkflowActivities is an interface for interacting with selected workflow activities.
	WorkflowActivities interface {
		Activities

		// GetByID will select an activity for the selected workflow by the given activityID.
		GetByID(activityID string) Activity
	}

	Activity interface {
		// Complete reports activity completed.
		// Activity Execute method can return activity.ErrResultPending to
		// indicate the activity is not completed when it's Execute method returns. In that case, this Complete method
		// should be called when that activity is completed with the actual result and error. If err is nil, activity task
		// completed event will be reported; if err is CanceledError, activity task cancelled event will be reported; otherwise,
		// activity task failed event will be reported.
		Complete(ctx context.Context, result interface{}, err error, opts ...ActivityCompleteOption) error

		// RecordHeartbeat records heartbeat for an activity.
		// details - is the progress you want to record along with heart beat for this activity.
		RecordHeartbeat(ctx context.Context, details interface{}, opts ...ActivityHeartbeatOption) error
	}

	// TaskLists is an interface for interacting with task lists.
	TaskLists interface {
		// Get selects a task list by the given name and type.
		Get(name string, taskListType TaskListType) TaskList
	}

	// TaskList is a light-weight Cadence queue that is used to deliver tasks to Cadence worker.
	TaskList interface {
		// Describe returns information about the tasklist, right now this API returns the
		// pollers which polled this tasklist in last few minutes.
		Describe(ctx context.Context) (TaskListInfo, error)
	}

	// SearchAttributes is an interface for interacting with search attributes.
	// The search attributes can be used in query of workflow List/Count APIs.
	// Adding new search attributes requires cadence server to update dynamic config ValidSearchAttributes.
	SearchAttributes interface {
		// List returns all valid search attributes keys and value types for the Cadence cluster.
		List(ctx context.Context) ([]SearchAttribute, error)
	}
)

// Page contains pagination data
// Zero value Page{}, means first page with default size
type Page struct {
	Size  int32  // 0 - means default page size
	Token []byte // nil - means first page
}

// FirstPage returns zero value Page struct which can be used as first page with default page size
func FirstPage() Page {
	return Page{}
}

type (
	// DomainReplicationConfig configures domain replication. Use GlobalDomain or LocalDomain to create it.
	DomainReplicationConfig interface {
		domainReplicationConfig()
	}

	// DomainRegisterOption allows passing optional parameters when registering a domain
	DomainRegisterOption interface {
		domainRegisterOption()
	}

	// DomainUpdateOption allows passing optional parameters when updating the domain
	DomainUpdateOption interface {
		domainUpdateOption()
	}

	// DomainListOption allows passing optional parameters when listing domains
	DomainListOption interface {
		domainListOption()
	}

	// DomainDescribeOption allows passing optional parameters when describing domain
	DomainDescribeOption interface {
		domainDescribeOption()
	}

	// DomainFailoverOption allows passing optional parameters when making a domain failover
	DomainFailoverOption interface {
		domainFailoverOption()
	}

	// DomainDeprecateOption allows passing optional parameters when deprecating a domain
	DomainDeprecateOption interface {
		domainDeprecateOption()
	}

	// DomainAddBadBinaryOption allows passing optional parameters when adding bad binaries
	DomainAddBadBinaryOption interface {
		domainAddBadBinaryOption()
	}

	// DomainDeleteBadBinaryOption allows passing optional parameters when deleting bad binaries
	DomainDeleteBadBinaryOption interface {
		domainDeleteBadBinaryOption()
	}

	// DomainOption can be used when registering or updating a domain
	DomainOption interface {
		DomainRegisterOption
		DomainUpdateOption
	}

	// WorkflowStartOption allows passing optional parameters when starting a workflow.
	WorkflowStartOption interface {
		workflowStartOption()
	}

	// WorkflowSignalOption allows passing optional parameters when signaling a workflow.
	WorkflowSignalOption interface {
		workflowSignalOption()
	}

	// WorkflowQueryOption allows passing optional parameters when querying a workflow.
	WorkflowQueryOption interface {
		workflowQueryOption()
	}

	// WorkflowDescribeOption allows passing optional parameters when describing a workflow.
	WorkflowDescribeOption interface {
		workflowDescribeOption()
	}

	// WorkflowCancelOption allows passing optional parameters when canceling a workflow.
	WorkflowCancelOption interface {
		workflowCancelOption()
	}

	// WorkflowTerminateOption allows passing optional parameters when terminating a workflow.
	WorkflowTerminateOption interface {
		workflowTerminateOption()
	}

	// WorkflowResetOption allows passing optional parameters when resetting a workflow.
	WorkflowResetOption interface {
		workflowResetOption()
	}

	// WorkflowGetResultOption allows passing optional parameters when retrieving workflow result.
	WorkflowGetResultOption interface {
		workflowGetResultOption()
	}

	// WorkflowCountOption allows passing optional parameters when counting workflows.
	WorkflowCountOption interface {
		workflowCountOption()
	}

	// WorkflowListOption allows passing optional parameters when listing workflows.
	WorkflowListOption interface {
		workflowListOption()
	}

	// ActivityCompleteOption allows passing optional parameters when asynchronously completing an activity.
	ActivityCompleteOption interface {
		activityCompleteOption()
	}

	// ActivityHeartbeatOption allows passing optional parameters when heartbeating an activity.
	ActivityHeartbeatOption interface {
		activityHeartbeatOption()
	}

	// WorkflowResetPoint specifies the point in history for workflow reset.
	WorkflowResetPoint interface {
		workflowResetPoint()
	}
)

// GlobalDomain can be used when registering a new domain to create its replication configuration.
// - clusters specify in which clusters domain will be registered
// - activeCluster specify an inital active cluster for the domain. This can later be changed with domain failover operation.
func GlobalDomain(activeCluster string, clusters ...string) DomainReplicationConfig {
	panic("not implemented")
}

// LocalDomain can be used when registering a new domain to create its replication configuration.
// LocalDomain will register domain only in one specified cluster and will not setup any replication.
func LocalDomain(cluster string) DomainReplicationConfig {
	panic("not implemented")
}

// SetDomainDescription will set description of the domain
func SetDomainDescription(description string) DomainOption {
	panic("not implemented")
}

// SetDomainOwnerEmail will set owner email of the domain
func SetDomainOwnerEmail(email string) DomainOption {
	panic("not implemented")
}

// SetWorkflowRetentionPeriod will set workflow execution retention period of the domain
func SetWorkflowRetentionPeriod(period time.Duration) DomainOption {
	panic("not implemented")
}

// SetDomainData will set arbitrary data map provided by the user for the domain
func SetDomainData(data map[string]string) DomainOption {
	panic("not implemented")
}

// SetHistoryArchival will set history archival parameters for the domain
func SetHistoryArchival(status ArchivalStatus, uri string) DomainOption {
	panic("not implemented")
}

// SetVisibilityArchival will set visibility archival parameters for the domain
func SetVisibilityArchival(status ArchivalStatus, uri string) DomainOption {
	panic("not implemented")
}

// WithGracefulFailover will use a graceful domain failover with provided timeout
func WithGracefulFailover(timeout time.Duration) DomainFailoverOption {
	panic("not implemented")
}

// WithWorkflowID sets the business identifier of the workflow execution.
// Default: generated UUID
func WithWorkflowID(workflowID string) WorkflowStartOption {
	panic("not implemented")
}

// WithDecisionTaskStartToCloseTimeout sets the timeout for processing decision task from the time the worker
// pulled this task. If a decision task is lost, it is retried after this timeout.
// The resolution is seconds.
// Default: 10 seconds
func WithDecisionTaskStartToCloseTimeout(timeout time.Duration) WorkflowStartOption {
	panic("not implemented")
}

// WithWorkflowIDReusePolicy defines whether server allow reuse of workflow ID.
// Can be useful for dedup logic if set to WorkflowIdReusePolicyRejectDuplicate.
// Default: WorkflowIDReusePolicyAllowDuplicateFailedOnly
func WithWorkflowIDReusePolicy(policy WorkflowIDReusePolicy) WorkflowStartOption {
	panic("not implemented")
}

// WithRetryPolicy sets the retry policy for the workflow.
// If provided, in case of workflow failure server will start new workflow execution if needed based on the retry policy.
func WithRetryPolicy(policy RetryPolicy) WorkflowStartOption {
	panic("not implemented")
}

// WithCronSchedule sets the cron schedule for workflow. If a cron schedule is specified, the workflow will run
// as a cron based on the schedule. The scheduling will be based on UTC time. Schedule for next run only happen
// after the current run is completed/failed/timeout. If a RetryPolicy is also supplied, and the workflow failed
// or timeout, the workflow will be retried based on the retry policy. While the workflow is retrying, it won't
// schedule its next run. If next schedule is due while workflow is running (or retrying), then it will skip that
// schedule. Cron workflow will not stop until it is terminated or cancelled (by returning cadence.CanceledError).
// The cron spec is as following:
// ┌───────────── minute (0 - 59)
// │ ┌───────────── hour (0 - 23)
// │ │ ┌───────────── day of the month (1 - 31)
// │ │ │ ┌───────────── month (1 - 12)
// │ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday)
// │ │ │ │ │
// │ │ │ │ │
// * * * * *
func WithCronSchedule(cron string) WorkflowStartOption {
	panic("not implemented")
}

// WithMemo sets non-indexed info that will be shown in list workflow.
func WithMemo(memo map[string]interface{}) WorkflowStartOption {
	panic("not implemented")
}

// WithSearchAttributes sets indexed info that can be used in query of List/Count workflow APIs (only
// supported when Cadence server is using ElasticSearch). The key and value type must be registered on Cadence server side.
// Use client.SearchAttributes().List() to get valid keys and corresponding value types.
func WithSearchAttributes(searchAttributes map[string]interface{}) WorkflowStartOption {
	panic("not implemented")
}

// WithDelayedStart sets the delay of the workflow start.
// The resolution is seconds.
// Default: 0 seconds
func WithDelayedStart(delay time.Duration) WorkflowStartOption {
	panic("not implemented")
}

// WithStart will start a workflow if he workflow is not running or not found, it starts the workflow and then sends the signal in transaction.
func WithStart(wfFunc interface{}, args []interface{}, taskList string, timeout time.Duration, opts ...WorkflowStartOption) WorkflowSignalOption {
	panic("not implemented")
}

// WithQueryRejectCondition sets the query behaviour based on workflow state.
func WithQueryRejectCondition(condition QueryRejectCondition) WorkflowQueryOption {
	panic("not implemented")
}

// WithQueryConsistencyLevel sets the consistency level on query.
// Default: QueryConsistencyLevelEventual
func WithQueryConsistencyLevel(level QueryConsistencyLevel) WorkflowQueryOption {
	panic("not implemented")
}

// WithTerminateReason provides a details why workflows is being terminated.
func WithTerminateDetails(details []byte) WorkflowTerminateOption {
	panic("not implemented")
}

// WithNoSignalReapply will not apply signals after the reset point.
func WithNoSignalReapply() WorkflowResetOption {
	panic("not implemented")
}

// ResetToLastCompletedDecision will reset the workfow to the last completed decision.
func ResetToLastCompletedDecision() WorkflowResetPoint {
	panic("not implemented")
}

// ResetToLastScheduledDecision will reset the workflow to the last scheduled decision.
func ResetToLastScheduledDecision() WorkflowResetPoint {
	panic("not implemented")
}

// ResetToLastContinueAsNew will reset the workflow to the last continue-as-new.
func ResetToLastContinueAsNew() WorkflowResetPoint {
	panic("not implemented")
}

// ResetToFirstCompletedDecision will reset the workflow to the first completed decision.
func ResetToFirstCompletedDecision() WorkflowResetPoint {
	panic("not implemented")
}

// ResetToFirstScheduledDecision will reset the workflow to the first scheduled decision.
func ResetToFirstScheduledDecision() WorkflowResetPoint {
	panic("not implemented")
}

// ResetToBadBinary will reset the workflow to the given bad binary checksum.
func ResetToBadBinary(checksum string) WorkflowResetPoint {
	panic("not implemented")
}

// ResetToEarliestDecisionCompletedAfter will reset the workflow to the earliest completed decition after the given timestamp.
func ResetToEarliestDecisionCompletedAfter(timestamp time.Time) WorkflowResetPoint {
	panic("not implemented")
}

// ResetToEventId will reset the workflow to exact given event ID.
func ResetToEventId(eventId int64) WorkflowResetPoint {
	panic("not implemented")
}

type (
	// QueryBuilder allows building worklows queries
	QueryBuilder interface {
		// Raw accepts raw SQL query
		Raw(sql string) Query

		// WorkflowStart constructs query to match workflows on given workflow start time range.
		// If passed time is nil, it is treated as an open ended interval.
		WorkflowStart(from, to *time.Time) Query
		// WorkflowClose constructs query to match workflows on given workflow close time range.
		// If passed time is nil, it is treated as an open ended interval.
		WorkflowClose(from, to *time.Time) Query
		// WorkflowStatus constructs query to match workflows with status in the given collection.
		WorkflowStatus(in ...WorkflowStatus) Query
		// WorkflowType constructs query to match workflow with exact worfklow type.
		WorkflowType(wfType string) Query
		// SearchAttribute constructs query to match workflow on arbitraty search attribute.
		SearchAttribute(key string, value interface{}) Query

		// And combines two queries so that both of them has to match the workflow
		And(a, b Query) Query
		// Or combines two quries so that only one of them has to match the workflow
		Or(a, b Query) Query
		// Not inverts the given query to not match the workflow
		Not(a Query) Query

		// All combines many queries so that all of them has to match the workflow
		All(clauses ...Query) Query
		// Any combines many queries so that any of them has to match the workflow
		Any(clauses ...Query) Query
	}

	// Query is a query constructed via QueryBuilder that can be used to query workflows.
	Query interface {
		// Validate can check whether constructed query is valid without issuing it to Cadence server.
		Validate() error
	}
)
