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

package metrics

import (
	"sync"
	"time"

	"github.com/uber-go/tally"
	"github.com/uber/tchannel-go/thrift"
	m "go.uber.org/cadence/.gen/go/cadence"
	"go.uber.org/cadence/.gen/go/shared"
)

type workflowServiceMetricsWrapper struct {
	service     m.TChanWorkflowService
	scope       tally.Scope
	childScopes map[string]tally.Scope
	mutex       sync.Mutex
}

const (
	scopeNameDeprecateDomain                = "DeprecateDomain"
	scopeNameDescribeDomain                 = "DescribeDomain"
	scopeNameGetWorkflowExecutionHistory    = "GetWorkflowExecutionHistory"
	scopeNameListClosedWorkflowExecutions   = "ListClosedWorkflowExecutions"
	scopeNameListOpenWorkflowExecutions     = "ListOpenWorkflowExecutions"
	scopeNamePollForActivityTask            = "PollForActivityTask"
	scopeNamePollForDecisionTask            = "PollForDecisionTask"
	scopeNameRecordActivityTaskHeartbeat    = "RecordActivityTaskHeartbeat"
	scopeNameRegisterDomain                 = "RegisterDomain"
	scopeNameRequestCancelWorkflowExecution = "RequestCancelWorkflowExecution"
	scopeNameRespondActivityTaskCanceled    = "RespondActivityTaskCanceled"
	scopeNameRespondActivityTaskCompleted   = "RespondActivityTaskCompleted"
	scopeNameRespondActivityTaskFailed      = "RespondActivityTaskFailed"
	scopeNameRespondDecisionTaskCompleted   = "RespondDecisionTaskCompleted"
	scopeNameSignalWorkflowExecution        = "SignalWorkflowExecution"
	scopeNameStartWorkflowExecution         = "StartWorkflowExecution"
	scopeNameTerminateWorkflowExecution     = "TerminateWorkflowExecution"
	scopeNameUpdateDomain                   = "UpdateDomain"
)

// NewWorkflowServiceWrapper creates a new wrapper to WorkflowService that will emit metrics for each service call.
func NewWorkflowServiceWrapper(service m.TChanWorkflowService, scope tally.Scope) m.TChanWorkflowService {
	return &workflowServiceMetricsWrapper{service: service, scope: scope, childScopes: make(map[string]tally.Scope)}
}

func (w *workflowServiceMetricsWrapper) getScope(scopeName string) tally.Scope {
	w.mutex.Lock()
	scope, ok := w.childScopes[scopeName]
	if ok {
		w.mutex.Unlock()
		return scope
	}
	scope = w.scope.SubScope(scopeName)
	w.childScopes[scopeName] = scope
	w.mutex.Unlock()
	return scope
}

func (w *workflowServiceMetricsWrapper) DeprecateDomain(ctx thrift.Context, request *shared.DeprecateDomainRequest) error {
	scope := w.getScope(scopeNameDeprecateDomain)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	err := w.service.DeprecateDomain(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return err
}

func (w *workflowServiceMetricsWrapper) DescribeDomain(ctx thrift.Context, request *shared.DescribeDomainRequest) (*shared.DescribeDomainResponse, error) {
	scope := w.getScope(scopeNameDescribeDomain)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	result, err := w.service.DescribeDomain(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return result, err
}

func (w *workflowServiceMetricsWrapper) GetWorkflowExecutionHistory(ctx thrift.Context, request *shared.GetWorkflowExecutionHistoryRequest) (*shared.GetWorkflowExecutionHistoryResponse, error) {
	scope := w.getScope(scopeNameGetWorkflowExecutionHistory)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	result, err := w.service.GetWorkflowExecutionHistory(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return result, err
}

func (w *workflowServiceMetricsWrapper) ListClosedWorkflowExecutions(ctx thrift.Context, request *shared.ListClosedWorkflowExecutionsRequest) (*shared.ListClosedWorkflowExecutionsResponse, error) {
	scope := w.getScope(scopeNameListClosedWorkflowExecutions)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	result, err := w.service.ListClosedWorkflowExecutions(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return result, err
}

func (w *workflowServiceMetricsWrapper) ListOpenWorkflowExecutions(ctx thrift.Context, request *shared.ListOpenWorkflowExecutionsRequest) (*shared.ListOpenWorkflowExecutionsResponse, error) {
	scope := w.getScope(scopeNameListOpenWorkflowExecutions)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	result, err := w.service.ListOpenWorkflowExecutions(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return result, err
}

func (w *workflowServiceMetricsWrapper) PollForActivityTask(ctx thrift.Context, request *shared.PollForActivityTaskRequest) (*shared.PollForActivityTaskResponse, error) {
	scope := w.getScope(scopeNamePollForActivityTask)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	result, err := w.service.PollForActivityTask(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return result, err
}

func (w *workflowServiceMetricsWrapper) PollForDecisionTask(ctx thrift.Context, request *shared.PollForDecisionTaskRequest) (*shared.PollForDecisionTaskResponse, error) {
	scope := w.getScope(scopeNamePollForDecisionTask)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	result, err := w.service.PollForDecisionTask(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return result, err
}

func (w *workflowServiceMetricsWrapper) RecordActivityTaskHeartbeat(ctx thrift.Context, request *shared.RecordActivityTaskHeartbeatRequest) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	scope := w.getScope(scopeNameRecordActivityTaskHeartbeat)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	result, err := w.service.RecordActivityTaskHeartbeat(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return result, err
}

func (w *workflowServiceMetricsWrapper) RegisterDomain(ctx thrift.Context, request *shared.RegisterDomainRequest) error {
	scope := w.getScope(scopeNameRegisterDomain)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	err := w.service.RegisterDomain(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return err
}

func (w *workflowServiceMetricsWrapper) RequestCancelWorkflowExecution(ctx thrift.Context, request *shared.RequestCancelWorkflowExecutionRequest) error {
	scope := w.getScope(scopeNameRequestCancelWorkflowExecution)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	err := w.service.RequestCancelWorkflowExecution(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return err
}

func (w *workflowServiceMetricsWrapper) RespondActivityTaskCanceled(ctx thrift.Context, request *shared.RespondActivityTaskCanceledRequest) error {
	scope := w.getScope(scopeNameRespondActivityTaskCanceled)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	err := w.service.RespondActivityTaskCanceled(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return err
}

func (w *workflowServiceMetricsWrapper) RespondActivityTaskCompleted(ctx thrift.Context, request *shared.RespondActivityTaskCompletedRequest) error {
	scope := w.getScope(scopeNameRespondActivityTaskCompleted)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	err := w.service.RespondActivityTaskCompleted(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return err
}

func (w *workflowServiceMetricsWrapper) RespondActivityTaskFailed(ctx thrift.Context, request *shared.RespondActivityTaskFailedRequest) error {
	scope := w.getScope(scopeNameRespondActivityTaskFailed)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	err := w.service.RespondActivityTaskFailed(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return err
}

func (w *workflowServiceMetricsWrapper) RespondDecisionTaskCompleted(ctx thrift.Context, request *shared.RespondDecisionTaskCompletedRequest) error {
	scope := w.getScope(scopeNameRespondDecisionTaskCompleted)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	err := w.service.RespondDecisionTaskCompleted(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return err
}

func (w *workflowServiceMetricsWrapper) SignalWorkflowExecution(ctx thrift.Context, request *shared.SignalWorkflowExecutionRequest) error {
	scope := w.getScope(scopeNameSignalWorkflowExecution)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	err := w.service.SignalWorkflowExecution(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return err
}

func (w *workflowServiceMetricsWrapper) StartWorkflowExecution(ctx thrift.Context, request *shared.StartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse, error) {
	scope := w.getScope(scopeNameStartWorkflowExecution)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	result, err := w.service.StartWorkflowExecution(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return result, err
}

func (w *workflowServiceMetricsWrapper) TerminateWorkflowExecution(ctx thrift.Context, request *shared.TerminateWorkflowExecutionRequest) error {
	scope := w.getScope(scopeNameTerminateWorkflowExecution)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	err := w.service.TerminateWorkflowExecution(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return err
}

func (w *workflowServiceMetricsWrapper) UpdateDomain(ctx thrift.Context, request *shared.UpdateDomainRequest) (*shared.UpdateDomainResponse, error) {
	scope := w.getScope(scopeNameUpdateDomain)
	scope.Counter(CadenceRequest).Inc(1)
	startTime := time.Now()
	result, err := w.service.UpdateDomain(ctx, request)
	scope.Timer(CadenceLatency).Record(time.Now().Sub(startTime))
	if err != nil {
		scope.Counter(CadenceError).Inc(1)
	}
	return result, err
}
