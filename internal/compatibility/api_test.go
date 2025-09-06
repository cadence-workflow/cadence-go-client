// Copyright (c) 2021 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package compatibility

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	protobuf "github.com/gogo/protobuf/proto"
	gogo "github.com/gogo/protobuf/types"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/compatibility/proto"
	"go.uber.org/cadence/internal/compatibility/testdata"
	"go.uber.org/cadence/internal/compatibility/thrift"
	"go.uber.org/cadence/test/testdatagen"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

// Fuzzing configuration constants
const (
	// DefaultNilChance is the default probability of setting pointer/slice fields to nil
	DefaultNilChance = 0.25
	// DefaultIterations is the default number of fuzzing iterations to run
	DefaultIterations = 100
	// MaxSafeTimestampSeconds is the maximum seconds value that fits safely in int64 nanoseconds
	// This avoids overflow when converting to UnixNano (max safe range from 1970 to ~2262)
	MaxSafeTimestampSeconds = 9223372036
	// MaxDurationSeconds is the maximum duration in seconds that fits in int32
	// This prevents overflow in mapper conversion (durationToSeconds -> int32)
	// int32 max value is 2147483647, so use that as the safe limit
	MaxDurationSeconds = 2147483647
	// NanosecondsPerSecond is the number of nanoseconds in a second
	NanosecondsPerSecond = 1000000000
	// MaxPayloadBytes is the maximum payload size for fuzzing
	MaxPayloadBytes = 10
)

// FuzzOptions provides configuration for runFuzzTest
type FuzzOptions struct {
	// CustomFuncs are custom fuzzer functions to apply for specific types
	CustomFuncs []interface{}
	// ExcludedFields are field names to exclude from fuzzing (set to zero value)
	ExcludedFields []string
	// NilChance is the probability of setting pointer/slice fields to nil (default DefaultNilChance)
	NilChance float64
	// Iterations is the number of fuzzing iterations to run (default DefaultIterations)
	Iterations int
}

func TestActivityLocalDispatchInfo(t *testing.T) {
	for _, item := range []*apiv1.ActivityLocalDispatchInfo{nil, {}, &testdata.ActivityLocalDispatchInfo} {
		assert.Equal(t, item, proto.ActivityLocalDispatchInfo(thrift.ActivityLocalDispatchInfo(item)))
	}

	runFuzzTest(t,
		thrift.ActivityLocalDispatchInfo,
		proto.ActivityLocalDispatchInfo,
		FuzzOptions{},
	)
}
func TestActivityTaskCancelRequestedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskCancelRequestedEventAttributes{nil, {}, &testdata.ActivityTaskCancelRequestedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskCancelRequestedEventAttributes(thrift.ActivityTaskCancelRequestedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ActivityTaskCancelRequestedEventAttributes,
		proto.ActivityTaskCancelRequestedEventAttributes,
		FuzzOptions{},
	)
}
func TestActivityTaskCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskCanceledEventAttributes{nil, {}, &testdata.ActivityTaskCanceledEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskCanceledEventAttributes(thrift.ActivityTaskCanceledEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ActivityTaskCanceledEventAttributes,
		proto.ActivityTaskCanceledEventAttributes,
		FuzzOptions{},
	)
}
func TestActivityTaskCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskCompletedEventAttributes{nil, {}, &testdata.ActivityTaskCompletedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskCompletedEventAttributes(thrift.ActivityTaskCompletedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ActivityTaskCompletedEventAttributes,
		proto.ActivityTaskCompletedEventAttributes,
		FuzzOptions{},
	)
}
func TestActivityTaskFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskFailedEventAttributes{nil, {}, &testdata.ActivityTaskFailedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskFailedEventAttributes(thrift.ActivityTaskFailedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ActivityTaskFailedEventAttributes,
		proto.ActivityTaskFailedEventAttributes,
		FuzzOptions{},
	)
}
func TestActivityTaskScheduledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskScheduledEventAttributes{nil, {}, &testdata.ActivityTaskScheduledEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskScheduledEventAttributes(thrift.ActivityTaskScheduledEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ActivityTaskScheduledEventAttributes,
		proto.ActivityTaskScheduledEventAttributes,
		FuzzOptions{
			ExcludedFields: []string{
				"TaskList", // [TOO HARD] GoFuzz has issues with complex nested types, and TaskList is tested below in TestTaskList
			},
		},
	)
}
func TestActivityTaskStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskStartedEventAttributes{nil, {}, &testdata.ActivityTaskStartedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskStartedEventAttributes(thrift.ActivityTaskStartedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ActivityTaskStartedEventAttributes,
		proto.ActivityTaskStartedEventAttributes,
		FuzzOptions{},
	)
}
func TestActivityTaskTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskTimedOutEventAttributes{nil, {}, &testdata.ActivityTaskTimedOutEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskTimedOutEventAttributes(thrift.ActivityTaskTimedOutEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ActivityTaskTimedOutEventAttributes,
		proto.ActivityTaskTimedOutEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TimeoutType, c fuzz.Continue) {
					validValues := []apiv1.TimeoutType{
						apiv1.TimeoutType_TIMEOUT_TYPE_INVALID,
						apiv1.TimeoutType_TIMEOUT_TYPE_START_TO_CLOSE,
						apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_START,
						apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
						apiv1.TimeoutType_TIMEOUT_TYPE_HEARTBEAT,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestActivityType(t *testing.T) {
	for _, item := range []*apiv1.ActivityType{nil, {}, &testdata.ActivityType} {
		assert.Equal(t, item, proto.ActivityType(thrift.ActivityType(item)))
	}

	runFuzzTest(t,
		thrift.ActivityType,
		proto.ActivityType,
		FuzzOptions{},
	)
}
func TestBadBinaries(t *testing.T) {
	for _, item := range []*apiv1.BadBinaries{nil, {}, &testdata.BadBinaries} {
		assert.Equal(t, item, proto.BadBinaries(thrift.BadBinaries(item)))
	}

	runFuzzTest(t,
		thrift.BadBinaries,
		proto.BadBinaries,
		FuzzOptions{
			ExcludedFields: []string{
				"Binaries", // [TOO HARD] clearFieldsIf has issues with nested maps, and Binaries is tested below in TestBadBinaryInfo
			},
		},
	)
}
func TestBadBinaryInfo(t *testing.T) {
	for _, item := range []*apiv1.BadBinaryInfo{nil, {}, &testdata.BadBinaryInfo} {
		assert.Equal(t, item, proto.BadBinaryInfo(thrift.BadBinaryInfo(item)))
	}

	runFuzzTest(t,
		thrift.BadBinaryInfo,
		proto.BadBinaryInfo,
		FuzzOptions{},
	)
}
func TestCancelTimerFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.CancelTimerFailedEventAttributes{nil, {}, &testdata.CancelTimerFailedEventAttributes} {
		assert.Equal(t, item, proto.CancelTimerFailedEventAttributes(thrift.CancelTimerFailedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.CancelTimerFailedEventAttributes,
		proto.CancelTimerFailedEventAttributes,
		FuzzOptions{},
	)
}
func TestChildWorkflowExecutionCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionCanceledEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionCanceledEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionCanceledEventAttributes(thrift.ChildWorkflowExecutionCanceledEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ChildWorkflowExecutionCanceledEventAttributes,
		proto.ChildWorkflowExecutionCanceledEventAttributes,
		FuzzOptions{},
	)
}
func TestChildWorkflowExecutionCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionCompletedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionCompletedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionCompletedEventAttributes(thrift.ChildWorkflowExecutionCompletedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ChildWorkflowExecutionCompletedEventAttributes,
		proto.ChildWorkflowExecutionCompletedEventAttributes,
		FuzzOptions{},
	)
}
func TestChildWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionFailedEventAttributes(thrift.ChildWorkflowExecutionFailedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ChildWorkflowExecutionFailedEventAttributes,
		proto.ChildWorkflowExecutionFailedEventAttributes,
		FuzzOptions{},
	)
}
func TestChildWorkflowExecutionStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionStartedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionStartedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionStartedEventAttributes(thrift.ChildWorkflowExecutionStartedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ChildWorkflowExecutionStartedEventAttributes,
		proto.ChildWorkflowExecutionStartedEventAttributes,
		FuzzOptions{},
	)
}
func TestChildWorkflowExecutionTerminatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionTerminatedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionTerminatedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionTerminatedEventAttributes(thrift.ChildWorkflowExecutionTerminatedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ChildWorkflowExecutionTerminatedEventAttributes,
		proto.ChildWorkflowExecutionTerminatedEventAttributes,
		FuzzOptions{},
	)
}
func TestChildWorkflowExecutionTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionTimedOutEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionTimedOutEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionTimedOutEventAttributes(thrift.ChildWorkflowExecutionTimedOutEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ChildWorkflowExecutionTimedOutEventAttributes,
		proto.ChildWorkflowExecutionTimedOutEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TimeoutType, c fuzz.Continue) {
					validValues := []apiv1.TimeoutType{
						apiv1.TimeoutType_TIMEOUT_TYPE_INVALID,
						apiv1.TimeoutType_TIMEOUT_TYPE_START_TO_CLOSE,
						apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_START,
						apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
						apiv1.TimeoutType_TIMEOUT_TYPE_HEARTBEAT,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestClusterReplicationConfiguration(t *testing.T) {
	for _, item := range []*apiv1.ClusterReplicationConfiguration{nil, {}, &testdata.ClusterReplicationConfiguration} {
		assert.Equal(t, item, proto.ClusterReplicationConfiguration(thrift.ClusterReplicationConfiguration(item)))
	}

	runFuzzTest(t,
		thrift.ClusterReplicationConfiguration,
		proto.ClusterReplicationConfiguration,
		FuzzOptions{},
	)
}
func TestCountWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.CountWorkflowExecutionsRequest{nil, {}, &testdata.CountWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.CountWorkflowExecutionsRequest(thrift.CountWorkflowExecutionsRequest(item)))
	}

	runFuzzTest(t,
		thrift.CountWorkflowExecutionsRequest,
		proto.CountWorkflowExecutionsRequest,
		FuzzOptions{},
	)
}
func TestCountWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.CountWorkflowExecutionsResponse{nil, {}, &testdata.CountWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.CountWorkflowExecutionsResponse(thrift.CountWorkflowExecutionsResponse(item)))
	}

	runFuzzTest(t,
		thrift.CountWorkflowExecutionsResponse,
		proto.CountWorkflowExecutionsResponse,
		FuzzOptions{},
	)
}
func TestDataBlob(t *testing.T) {
	for _, item := range []*apiv1.DataBlob{nil, {}, &testdata.DataBlob} {
		assert.Equal(t, item, proto.DataBlob(thrift.DataBlob(item)))
	}

	runFuzzTest(t,
		thrift.DataBlob,
		proto.DataBlob,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.EncodingType, c fuzz.Continue) {
					validValues := []apiv1.EncodingType{
						apiv1.EncodingType_ENCODING_TYPE_INVALID,
						apiv1.EncodingType_ENCODING_TYPE_THRIFTRW,
						apiv1.EncodingType_ENCODING_TYPE_JSON,
						// TODO: Determine appropriate mapper behaviour for invalid values
						// ENCODING_TYPE_PROTO3 intentionally excluded as the mapper panics for this value
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestDecisionTaskCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskCompletedEventAttributes{nil, {}, &testdata.DecisionTaskCompletedEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskCompletedEventAttributes(thrift.DecisionTaskCompletedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.DecisionTaskCompletedEventAttributes,
		proto.DecisionTaskCompletedEventAttributes,
		FuzzOptions{},
	)
}
func TestDecisionTaskFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskFailedEventAttributes{nil, {}, &testdata.DecisionTaskFailedEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskFailedEventAttributes(thrift.DecisionTaskFailedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.DecisionTaskFailedEventAttributes,
		proto.DecisionTaskFailedEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(cause *apiv1.DecisionTaskFailedCause, c fuzz.Continue) {
					validValues := []apiv1.DecisionTaskFailedCause{
						apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_UNHANDLED_DECISION,
						apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SCHEDULE_ACTIVITY_ATTRIBUTES,
						apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_ACTIVITY_ATTRIBUTES,
						apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_START_TIMER_ATTRIBUTES,
						apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_TIMER_ATTRIBUTES,
					}
					*cause = validValues[c.Intn(len(validValues))]
				},
			},
			ExcludedFields: []string{
				"RequestId", // [BUG] Excluded as RequestId is currently not mapped.
			},
		},
	)
}
func TestDecisionTaskScheduledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskScheduledEventAttributes{nil, {}, &testdata.DecisionTaskScheduledEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskScheduledEventAttributes(thrift.DecisionTaskScheduledEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.DecisionTaskScheduledEventAttributes,
		proto.DecisionTaskScheduledEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TaskListKind, c fuzz.Continue) {
					validValues := []apiv1.TaskListKind{
						apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
						apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
						apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestDecisionTaskStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskStartedEventAttributes{nil, {}, &testdata.DecisionTaskStartedEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskStartedEventAttributes(thrift.DecisionTaskStartedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.DecisionTaskStartedEventAttributes,
		proto.DecisionTaskStartedEventAttributes,
		FuzzOptions{},
	)
}
func TestDecisionTaskTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskTimedOutEventAttributes{nil, {}, &testdata.DecisionTaskTimedOutEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskTimedOutEventAttributes(thrift.DecisionTaskTimedOutEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.DecisionTaskTimedOutEventAttributes,
		proto.DecisionTaskTimedOutEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(cause *apiv1.DecisionTaskTimedOutCause, c fuzz.Continue) {
					validValues := []apiv1.DecisionTaskTimedOutCause{
						apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_TIMEOUT,
						apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_RESET,
					}
					*cause = validValues[c.Intn(len(validValues))]
				},
				func(timeoutType *apiv1.TimeoutType, c fuzz.Continue) {
					validValues := []apiv1.TimeoutType{
						apiv1.TimeoutType_TIMEOUT_TYPE_START_TO_CLOSE,
						apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_START,
						apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
						apiv1.TimeoutType_TIMEOUT_TYPE_HEARTBEAT,
					}
					*timeoutType = validValues[c.Intn(len(validValues))]
				},
			},
			ExcludedFields: []string{
				"RequestId", // [BUG] Excluded as RequestId is currently not mapped.
			},
		},
	)
}
func TestDeprecateDomainRequest(t *testing.T) {
	for _, item := range []*apiv1.DeprecateDomainRequest{nil, {}, &testdata.DeprecateDomainRequest} {
		assert.Equal(t, item, proto.DeprecateDomainRequest(thrift.DeprecateDomainRequest(item)))
	}

	runFuzzTest(t,
		thrift.DeprecateDomainRequest,
		proto.DeprecateDomainRequest,
		FuzzOptions{},
	)
}
func TestDescribeDomainRequest(t *testing.T) {
	for _, item := range []*apiv1.DescribeDomainRequest{
		&testdata.DescribeDomainRequest_ID,
		&testdata.DescribeDomainRequest_Name,
	} {
		assert.Equal(t, item, proto.DescribeDomainRequest(thrift.DescribeDomainRequest(item)))
	}
	assert.Nil(t, proto.DescribeDomainRequest(nil))
	assert.Nil(t, thrift.DescribeDomainRequest(nil))
	assert.Panics(t, func() { proto.DescribeDomainRequest(&shared.DescribeDomainRequest{}) })
	assert.Panics(t, func() { thrift.DescribeDomainRequest(&apiv1.DescribeDomainRequest{}) })

	runFuzzTest(t,
		thrift.DescribeDomainRequest,
		proto.DescribeDomainRequest,
		FuzzOptions{
			CustomFuncs: []interface{}{
				// Custom fuzzer for DescribeDomainRequest to handle oneof interface
				func(d *apiv1.DescribeDomainRequest, c fuzz.Continue) {
					if c.RandBool() {
						d.DescribeBy = &apiv1.DescribeDomainRequest_Id{Id: c.RandString()}
					} else {
						d.DescribeBy = &apiv1.DescribeDomainRequest_Name{Name: c.RandString()}
					}
				},
			},
		},
	)
}
func TestDescribeDomainResponse_Domain(t *testing.T) {
	for _, item := range []*apiv1.Domain{nil, &testdata.Domain} {
		assert.Equal(t, item, proto.DescribeDomainResponseDomain(thrift.DescribeDomainResponseDomain(item)))
	}

	runFuzzTest(t,
		thrift.DescribeDomainResponseDomain,
		proto.DescribeDomainResponseDomain,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(status *apiv1.DomainStatus, c fuzz.Continue) {
					validValues := []apiv1.DomainStatus{
						apiv1.DomainStatus_DOMAIN_STATUS_INVALID,
						apiv1.DomainStatus_DOMAIN_STATUS_REGISTERED,
						apiv1.DomainStatus_DOMAIN_STATUS_DEPRECATED,
						apiv1.DomainStatus_DOMAIN_STATUS_DELETED,
					}
					*status = validValues[c.Intn(len(validValues))]
				},
				func(status *apiv1.ArchivalStatus, c fuzz.Continue) {
					validValues := []apiv1.ArchivalStatus{
						apiv1.ArchivalStatus_ARCHIVAL_STATUS_INVALID,
						apiv1.ArchivalStatus_ARCHIVAL_STATUS_DISABLED,
						apiv1.ArchivalStatus_ARCHIVAL_STATUS_ENABLED,
					}
					*status = validValues[c.Intn(len(validValues))]
				},
				// Custom fuzzer for WorkflowExecutionRetentionPeriod - must be day-precision
				// because thrift mapping uses durationToDays (truncates to day boundaries)
				func(domain *apiv1.Domain, c fuzz.Continue) {
					if domain.WorkflowExecutionRetentionPeriod != nil {
						// Generate days within int32 range: max ~5.8 million days (~16000 years)
						days := c.Int63n(MaxDurationSeconds / (24 * 3600))
						domain.WorkflowExecutionRetentionPeriod.Seconds = days * 24 * 3600
						domain.WorkflowExecutionRetentionPeriod.Nanos = 0
					}
				},
			},
			ExcludedFields: []string{},
		},
	)
}
func TestDescribeDomainResponse(t *testing.T) {
	for _, item := range []*apiv1.DescribeDomainResponse{nil, &testdata.DescribeDomainResponse} {
		assert.Equal(t, item, proto.DescribeDomainResponse(thrift.DescribeDomainResponse(item)))
	}

	runFuzzTest(t,
		thrift.DescribeDomainResponse,
		proto.DescribeDomainResponse,
		FuzzOptions{
			NilChance: 0.25,
			CustomFuncs: []interface{}{
				// TODO: Figure out this test - it seems to panic immediately
				func(resp *apiv1.DescribeDomainResponse, c fuzz.Continue) {
					resp.Domain = &apiv1.Domain{
						Name:                     c.RandString(),
						Status:                   apiv1.DomainStatus_DOMAIN_STATUS_REGISTERED,
						HistoryArchivalStatus:    apiv1.ArchivalStatus_ARCHIVAL_STATUS_DISABLED,
						VisibilityArchivalStatus: apiv1.ArchivalStatus_ARCHIVAL_STATUS_DISABLED,
					}
				},
			},
			ExcludedFields: []string{},
		},
	)
}
func TestDescribeTaskListRequest(t *testing.T) {
	for _, item := range []*apiv1.DescribeTaskListRequest{nil, {}, &testdata.DescribeTaskListRequest} {
		assert.Equal(t, item, proto.DescribeTaskListRequest(thrift.DescribeTaskListRequest(item)))
	}

	runFuzzTest(t,
		thrift.DescribeTaskListRequest,
		proto.DescribeTaskListRequest,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TaskListKind, c fuzz.Continue) {
					validValues := []apiv1.TaskListKind{
						apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
						apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
						apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
				func(e *apiv1.TaskListType, c fuzz.Continue) {
					validValues := []apiv1.TaskListType{
						apiv1.TaskListType_TASK_LIST_TYPE_INVALID,
						apiv1.TaskListType_TASK_LIST_TYPE_DECISION,
						apiv1.TaskListType_TASK_LIST_TYPE_ACTIVITY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestDescribeTaskListResponse(t *testing.T) {
	for _, item := range []*apiv1.DescribeTaskListResponse{nil, {}, &testdata.DescribeTaskListResponse} {
		assert.Equal(t, item, proto.DescribeTaskListResponse(thrift.DescribeTaskListResponse(item)))
	}

	runFuzzTest(t,
		thrift.DescribeTaskListResponse,
		proto.DescribeTaskListResponse,
		FuzzOptions{
			ExcludedFields: []string{
				"PartitionConfig", // [BUG] PartitionConfig field is lost during round trip - complex nested maps with TaskListPartition not preserved
				"TaskListStatus",  // [BUG] TaskListStatus fields IsolationGroupMetrics and NewTasksPerSecond are not mapped - they become nil/0 after round trip
				"TaskList",        // [BUG] TaskList field is lost during round trip - mapper not preserving this field correctly
			},
		},
	)
}
func TestDescribeWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.DescribeWorkflowExecutionRequest{nil, {}, &testdata.DescribeWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.DescribeWorkflowExecutionRequest(thrift.DescribeWorkflowExecutionRequest(item)))
	}

	runFuzzTest(t,
		thrift.DescribeWorkflowExecutionRequest,
		proto.DescribeWorkflowExecutionRequest,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.QueryConsistencyLevel, c fuzz.Continue) {
					validValues := []apiv1.QueryConsistencyLevel{
						apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_INVALID,
						apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_EVENTUAL,
						apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_STRONG,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestDescribeWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.DescribeWorkflowExecutionResponse{nil, {}, &testdata.DescribeWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.DescribeWorkflowExecutionResponse(thrift.DescribeWorkflowExecutionResponse(item)))
	}

	runFuzzTest(t,
		thrift.DescribeWorkflowExecutionResponse,
		proto.DescribeWorkflowExecutionResponse,
		FuzzOptions{
			CustomFuncs: []interface{}{
				// TODO: Figure out this test - it seems to panic immediately
				// Custom fuzzer to avoid gofuzz issues with complex types
				func(resp *apiv1.DescribeWorkflowExecutionResponse, c fuzz.Continue) {
					// Only populate simple fields to avoid gofuzz panics
					resp.ExecutionConfiguration = &apiv1.WorkflowExecutionConfiguration{
						TaskList: &apiv1.TaskList{Name: c.RandString()},
					}
				},
			},
			ExcludedFields: []string{},
		},
	)
}
func TestDiagnoseWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.DiagnoseWorkflowExecutionRequest{nil, {}, &testdata.DiagnoseWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.DiagnoseWorkflowExecutionRequest(thrift.DiagnoseWorkflowExecutionRequest(item)))
	}

	runFuzzTest(t,
		thrift.DiagnoseWorkflowExecutionRequest,
		proto.DiagnoseWorkflowExecutionRequest,
		FuzzOptions{},
	)
}
func TestDiagnoseWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.DiagnoseWorkflowExecutionResponse{nil, {}, &testdata.DiagnoseWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.DiagnoseWorkflowExecutionResponse(thrift.DiagnoseWorkflowExecutionResponse(item)))
	}

	runFuzzTest(t,
		thrift.DiagnoseWorkflowExecutionResponse,
		proto.DiagnoseWorkflowExecutionResponse,
		FuzzOptions{},
	)
}
func TestExternalWorkflowExecutionCancelRequestedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ExternalWorkflowExecutionCancelRequestedEventAttributes{nil, {}, &testdata.ExternalWorkflowExecutionCancelRequestedEventAttributes} {
		assert.Equal(t, item, proto.ExternalWorkflowExecutionCancelRequestedEventAttributes(thrift.ExternalWorkflowExecutionCancelRequestedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ExternalWorkflowExecutionCancelRequestedEventAttributes,
		proto.ExternalWorkflowExecutionCancelRequestedEventAttributes,
		FuzzOptions{},
	)
}
func TestExternalWorkflowExecutionSignaledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ExternalWorkflowExecutionSignaledEventAttributes{nil, {}, &testdata.ExternalWorkflowExecutionSignaledEventAttributes} {
		assert.Equal(t, item, proto.ExternalWorkflowExecutionSignaledEventAttributes(thrift.ExternalWorkflowExecutionSignaledEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.ExternalWorkflowExecutionSignaledEventAttributes,
		proto.ExternalWorkflowExecutionSignaledEventAttributes,
		FuzzOptions{},
	)
}
func TestGetClusterInfoResponse(t *testing.T) {
	for _, item := range []*apiv1.GetClusterInfoResponse{nil, {}, &testdata.GetClusterInfoResponse} {
		assert.Equal(t, item, proto.GetClusterInfoResponse(thrift.GetClusterInfoResponse(item)))
	}

	runFuzzTest(t,
		thrift.GetClusterInfoResponse,
		proto.GetClusterInfoResponse,
		FuzzOptions{},
	)
}
func TestGetSearchAttributesResponse(t *testing.T) {
	for _, item := range []*apiv1.GetSearchAttributesResponse{nil, {}, &testdata.GetSearchAttributesResponse} {
		assert.Equal(t, item, proto.GetSearchAttributesResponse(thrift.GetSearchAttributesResponse(item)))
	}

	runFuzzTest(t,
		thrift.GetSearchAttributesResponse,
		proto.GetSearchAttributesResponse,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.IndexedValueType, c fuzz.Continue) {
					// TODO: Support INDEXED_VALUE_TYPE_INVALID
					validValues := []apiv1.IndexedValueType{
						apiv1.IndexedValueType_INDEXED_VALUE_TYPE_STRING,
						apiv1.IndexedValueType_INDEXED_VALUE_TYPE_KEYWORD,
						apiv1.IndexedValueType_INDEXED_VALUE_TYPE_INT,
						apiv1.IndexedValueType_INDEXED_VALUE_TYPE_DOUBLE,
						apiv1.IndexedValueType_INDEXED_VALUE_TYPE_BOOL,
						apiv1.IndexedValueType_INDEXED_VALUE_TYPE_DATETIME,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestGetWorkflowExecutionHistoryRequest(t *testing.T) {
	for _, item := range []*apiv1.GetWorkflowExecutionHistoryRequest{nil, {}, &testdata.GetWorkflowExecutionHistoryRequest} {
		assert.Equal(t, item, proto.GetWorkflowExecutionHistoryRequest(thrift.GetWorkflowExecutionHistoryRequest(item)))
	}

	runFuzzTest(t,
		thrift.GetWorkflowExecutionHistoryRequest,
		proto.GetWorkflowExecutionHistoryRequest,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.EventFilterType, c fuzz.Continue) {
					validValues := []apiv1.EventFilterType{
						apiv1.EventFilterType_EVENT_FILTER_TYPE_INVALID,
						apiv1.EventFilterType_EVENT_FILTER_TYPE_ALL_EVENT,
						apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
				func(e *apiv1.QueryConsistencyLevel, c fuzz.Continue) {
					validValues := []apiv1.QueryConsistencyLevel{
						apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_INVALID,
						apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_EVENTUAL,
						apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_STRONG,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestGetWorkflowExecutionHistoryResponse(t *testing.T) {
	for _, item := range []*apiv1.GetWorkflowExecutionHistoryResponse{nil, {}, &testdata.GetWorkflowExecutionHistoryResponse} {
		assert.Equal(t, item, proto.GetWorkflowExecutionHistoryResponse(thrift.GetWorkflowExecutionHistoryResponse(item)))
	}

	runFuzzTest(t,
		thrift.GetWorkflowExecutionHistoryResponse,
		proto.GetWorkflowExecutionHistoryResponse,
		FuzzOptions{
			NilChance: 0.0, // Avoid gofuzz nil issues
			// TODO: Figure out this test - it seems to panic immediately
			CustomFuncs: []interface{}{
				// Custom fuzzer to avoid gofuzz panic with complex types
				func(resp *apiv1.GetWorkflowExecutionHistoryResponse, c fuzz.Continue) {
					// Only fuzz simple fields to avoid gofuzz panic
					resp.NextPageToken = make([]byte, c.Intn(10))
					for i := range resp.NextPageToken {
						resp.NextPageToken[i] = byte(c.Uint32())
					}
					resp.Archived = c.RandBool()
				},
				func(e *apiv1.EncodingType, c fuzz.Continue) {
					validValues := []apiv1.EncodingType{
						apiv1.EncodingType_ENCODING_TYPE_INVALID,
						apiv1.EncodingType_ENCODING_TYPE_THRIFTRW,
						apiv1.EncodingType_ENCODING_TYPE_JSON,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
			ExcludedFields: []string{
				"History", // Complex nested structure that causes gofuzz issues
			},
		},
	)
}
func TestHeader(t *testing.T) {
	for _, item := range []*apiv1.Header{nil, {}, &testdata.Header} {
		assert.Equal(t, item, proto.Header(thrift.Header(item)))
	}

	runFuzzTest(t,
		thrift.Header,
		proto.Header,
		FuzzOptions{},
	)
}
func TestHistory(t *testing.T) {
	for _, item := range []*apiv1.History{nil, {}, &testdata.History} {
		assert.Equal(t, item, proto.History(thrift.History(item)))
	}

	// [TOO HARD] HistoryEvents are too complex (particularly oneofs) for gofuzz to handle.
	// The events themselves are tested in TestHistoryEvent, and don't need to be fuzzed here.
}
func TestListArchivedWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListArchivedWorkflowExecutionsRequest{nil, {}, &testdata.ListArchivedWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.ListArchivedWorkflowExecutionsRequest(thrift.ListArchivedWorkflowExecutionsRequest(item)))
	}

	runFuzzTest(t,
		thrift.ListArchivedWorkflowExecutionsRequest,
		proto.ListArchivedWorkflowExecutionsRequest,
		FuzzOptions{},
	)
}
func TestListArchivedWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListArchivedWorkflowExecutionsResponse{nil, {}, &testdata.ListArchivedWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListArchivedWorkflowExecutionsResponse(thrift.ListArchivedWorkflowExecutionsResponse(item)))
	}

	runFuzzTest(t,
		thrift.ListArchivedWorkflowExecutionsResponse,
		proto.ListArchivedWorkflowExecutionsResponse,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(resp *apiv1.ListArchivedWorkflowExecutionsResponse, c fuzz.Continue) {
					c.Fuzz(&resp.NextPageToken)
				},
			},
			ExcludedFields: []string{
				"Executions", // [TOO HARD] Array of complex WorkflowExecutionInfo structures - tested in TestWorkflowExecutionInfo
			},
		},
	)
}
func TestListClosedWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListClosedWorkflowExecutionsResponse{nil, {}, &testdata.ListClosedWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListClosedWorkflowExecutionsResponse(thrift.ListClosedWorkflowExecutionsResponse(item)))
	}

	runFuzzTest(t,
		thrift.ListClosedWorkflowExecutionsResponse,
		proto.ListClosedWorkflowExecutionsResponse,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(resp *apiv1.ListClosedWorkflowExecutionsResponse, c fuzz.Continue) {
					c.Fuzz(&resp.NextPageToken)
				},
			},
			ExcludedFields: []string{
				"Executions", // [TOO HARD] Array of complex WorkflowExecutionInfo structures - tested in TestWorkflowExecutionInfo
			},
		},
	)
}
func TestListDomainsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListDomainsRequest{nil, {}, &testdata.ListDomainsRequest} {
		assert.Equal(t, item, proto.ListDomainsRequest(thrift.ListDomainsRequest(item)))
	}

	runFuzzTest(t,
		thrift.ListDomainsRequest,
		proto.ListDomainsRequest,
		FuzzOptions{
			// TODO: Re-enable NilChance and fix the mapper
			NilChance: 0.0,
			CustomFuncs: []interface{}{
				func(req *apiv1.ListDomainsRequest, c fuzz.Continue) {
					req.PageSize = c.Int31n(100)
					req.NextPageToken = make([]byte, c.Intn(10))
					for i := range req.NextPageToken {
						req.NextPageToken[i] = byte(c.Uint32())
					}
				},
			},
		},
	)
}
func TestListDomainsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListDomainsResponse{nil, {}, &testdata.ListDomainsResponse} {
		assert.Equal(t, item, proto.ListDomainsResponse(thrift.ListDomainsResponse(item)))
	}

	runFuzzTest(t,
		thrift.ListDomainsResponse,
		proto.ListDomainsResponse,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(status *apiv1.DomainStatus, c fuzz.Continue) {
					validValues := []apiv1.DomainStatus{
						apiv1.DomainStatus_DOMAIN_STATUS_INVALID,
						apiv1.DomainStatus_DOMAIN_STATUS_REGISTERED,
						apiv1.DomainStatus_DOMAIN_STATUS_DEPRECATED,
						apiv1.DomainStatus_DOMAIN_STATUS_DELETED,
					}
					*status = validValues[c.Intn(len(validValues))]
				},
				func(status *apiv1.ArchivalStatus, c fuzz.Continue) {
					validValues := []apiv1.ArchivalStatus{
						apiv1.ArchivalStatus_ARCHIVAL_STATUS_INVALID,
						apiv1.ArchivalStatus_ARCHIVAL_STATUS_DISABLED,
						apiv1.ArchivalStatus_ARCHIVAL_STATUS_ENABLED,
					}
					*status = validValues[c.Intn(len(validValues))]
				},
				// WorkflowExecutionRetentionPeriod - must be day-precision
				// because thrift mapping uses durationToDays (truncates to day boundaries)
				func(domain *apiv1.Domain, c fuzz.Continue) {
					if domain.WorkflowExecutionRetentionPeriod != nil {
						// Generate days within int32 range: max ~5.8 million days (~16000 years)
						days := c.Int63n(MaxDurationSeconds / (24 * 3600))
						domain.WorkflowExecutionRetentionPeriod.Seconds = days * 24 * 3600
						domain.WorkflowExecutionRetentionPeriod.Nanos = 0
					}
				},
			},
		},
	)
}
func TestListOpenWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListOpenWorkflowExecutionsResponse{nil, {}, &testdata.ListOpenWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListOpenWorkflowExecutionsResponse(thrift.ListOpenWorkflowExecutionsResponse(item)))
	}

	runFuzzTest(t,
		thrift.ListOpenWorkflowExecutionsResponse,
		proto.ListOpenWorkflowExecutionsResponse,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(resp *apiv1.ListOpenWorkflowExecutionsResponse, c fuzz.Continue) {
					c.Fuzz(&resp.NextPageToken)
				},
			},
			ExcludedFields: []string{
				"Executions", // [TOO HARD] Array of complex WorkflowExecutionInfo structures - tested in TestWorkflowExecutionInfo
			},
		},
	)
}
func TestListTaskListPartitionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListTaskListPartitionsRequest{nil, {}, &testdata.ListTaskListPartitionsRequest} {
		assert.Equal(t, item, proto.ListTaskListPartitionsRequest(thrift.ListTaskListPartitionsRequest(item)))
	}

	runFuzzTest(t,
		thrift.ListTaskListPartitionsRequest,
		proto.ListTaskListPartitionsRequest,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TaskListType, c fuzz.Continue) {
					validValues := []apiv1.TaskListType{
						apiv1.TaskListType_TASK_LIST_TYPE_INVALID,
						apiv1.TaskListType_TASK_LIST_TYPE_DECISION,
						apiv1.TaskListType_TASK_LIST_TYPE_ACTIVITY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
				func(e *apiv1.TaskListKind, c fuzz.Continue) {
					validValues := []apiv1.TaskListKind{
						apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
						apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
						apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestListTaskListPartitionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListTaskListPartitionsResponse{nil, {}, &testdata.ListTaskListPartitionsResponse} {
		assert.Equal(t, item, proto.ListTaskListPartitionsResponse(thrift.ListTaskListPartitionsResponse(item)))
	}

	runFuzzTest(t,
		thrift.ListTaskListPartitionsResponse,
		proto.ListTaskListPartitionsResponse,
		FuzzOptions{},
	)
}
func TestListWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListWorkflowExecutionsRequest{nil, {}, &testdata.ListWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.ListWorkflowExecutionsRequest(thrift.ListWorkflowExecutionsRequest(item)))
	}

	runFuzzTest(t,
		thrift.ListWorkflowExecutionsRequest,
		proto.ListWorkflowExecutionsRequest,
		FuzzOptions{},
	)
}
func TestListWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListWorkflowExecutionsResponse{nil, {}, &testdata.ListWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListWorkflowExecutionsResponse(thrift.ListWorkflowExecutionsResponse(item)))
	}

	runFuzzTest(t,
		thrift.ListWorkflowExecutionsResponse,
		proto.ListWorkflowExecutionsResponse,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(resp *apiv1.ListWorkflowExecutionsResponse, c fuzz.Continue) {
					c.Fuzz(&resp.NextPageToken)
				},
			},
			ExcludedFields: []string{
				"Executions", // [TOO HARD] Array of complex WorkflowExecutionInfo structures - tested in TestWorkflowExecutionInfo
			},
		},
	)
}
func TestMarkerRecordedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.MarkerRecordedEventAttributes{nil, {}, &testdata.MarkerRecordedEventAttributes} {
		assert.Equal(t, item, proto.MarkerRecordedEventAttributes(thrift.MarkerRecordedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.MarkerRecordedEventAttributes,
		proto.MarkerRecordedEventAttributes,
		FuzzOptions{},
	)
}
func TestMemo(t *testing.T) {
	for _, item := range []*apiv1.Memo{nil, {}, &testdata.Memo} {
		assert.Equal(t, item, proto.Memo(thrift.Memo(item)))
	}

	runFuzzTest(t,
		thrift.Memo,
		proto.Memo,
		FuzzOptions{},
	)
}
func TestPendingActivityInfo(t *testing.T) {
	for _, item := range []*apiv1.PendingActivityInfo{nil, {}, &testdata.PendingActivityInfo} {
		assert.Equal(t, item, proto.PendingActivityInfo(thrift.PendingActivityInfo(item)))
	}

	runFuzzTest(t,
		thrift.PendingActivityInfo,
		proto.PendingActivityInfo,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(state *apiv1.PendingActivityState, c fuzz.Continue) {
					validValues := []apiv1.PendingActivityState{
						apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_INVALID,
						apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_SCHEDULED,
						apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_STARTED,
						apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_CANCEL_REQUESTED,
					}
					*state = validValues[c.Intn(len(validValues))]
				},
			},
			ExcludedFields: []string{
				"StartedWorkerIdentity", // [BUG] StartedWorkerIdentity is not mapped
				"ScheduleId",            // [BUG] ScheduleId is not mapped
			},
		},
	)
}
func TestPendingChildExecutionInfo(t *testing.T) {
	for _, item := range []*apiv1.PendingChildExecutionInfo{nil, {}, &testdata.PendingChildExecutionInfo} {
		assert.Equal(t, item, proto.PendingChildExecutionInfo(thrift.PendingChildExecutionInfo(item)))
	}

	runFuzzTest(t,
		thrift.PendingChildExecutionInfo,
		proto.PendingChildExecutionInfo,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(policy *apiv1.ParentClosePolicy, c fuzz.Continue) {
					validValues := []apiv1.ParentClosePolicy{
						apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_INVALID,
						apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_ABANDON,
						apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_REQUEST_CANCEL,
						apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_TERMINATE,
					}
					*policy = validValues[c.Intn(len(validValues))]
				},
			},
			ExcludedFields: []string{
				"Domain", // [BUG] It is not clear why Domain is not mapped
			},
		},
	)
}
func TestPendingDecisionInfo(t *testing.T) {
	for _, item := range []*apiv1.PendingDecisionInfo{nil, {}, &testdata.PendingDecisionInfo} {
		assert.Equal(t, item, proto.PendingDecisionInfo(thrift.PendingDecisionInfo(item)))
	}

	runFuzzTest(t,
		thrift.PendingDecisionInfo,
		proto.PendingDecisionInfo,
		FuzzOptions{
			CustomFuncs: []interface{}{
				// PendingDecisionState has inconsistent behaviour between the proto and thrift mappers
				// proto will panic when INVALID is specified, but return INVALID when receiving nil
				// thrift returns nil when receiving INVALID, but panics for all other values
				// [BUG] TODO: Make the mappers consistent
				func(state *apiv1.PendingDecisionState, c fuzz.Continue) {
					validValues := []apiv1.PendingDecisionState{
						apiv1.PendingDecisionState_PENDING_DECISION_STATE_SCHEDULED,
						apiv1.PendingDecisionState_PENDING_DECISION_STATE_STARTED,
					}
					*state = validValues[c.Intn(len(validValues))]
				},
			},
			ExcludedFields: []string{
				"ScheduleId", // [BUG] ScheduleId is unmapped by either mapper function - it is not clear if this is intentional
			},
		},
	)
}
func TestPollForActivityTaskRequest(t *testing.T) {
	for _, item := range []*apiv1.PollForActivityTaskRequest{nil, {}, &testdata.PollForActivityTaskRequest} {
		assert.Equal(t, item, proto.PollForActivityTaskRequest(thrift.PollForActivityTaskRequest(item)))
	}

	runFuzzTest(t,
		thrift.PollForActivityTaskRequest,
		proto.PollForActivityTaskRequest,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TaskListKind, c fuzz.Continue) {
					validValues := []apiv1.TaskListKind{
						apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
						apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
						apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestPollForActivityTaskResponse(t *testing.T) {
	for _, item := range []*apiv1.PollForActivityTaskResponse{nil, {}, &testdata.PollForActivityTaskResponse} {
		assert.Equal(t, item, proto.PollForActivityTaskResponse(thrift.PollForActivityTaskResponse(item)))
	}

	runFuzzTest(t,
		thrift.PollForActivityTaskResponse,
		proto.PollForActivityTaskResponse,
		FuzzOptions{},
	)
}
func TestPollForDecisionTaskRequest(t *testing.T) {
	for _, item := range []*apiv1.PollForDecisionTaskRequest{nil, {}, &testdata.PollForDecisionTaskRequest} {
		assert.Equal(t, item, proto.PollForDecisionTaskRequest(thrift.PollForDecisionTaskRequest(item)))
	}

	runFuzzTest(t,
		thrift.PollForDecisionTaskRequest,
		proto.PollForDecisionTaskRequest,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TaskListKind, c fuzz.Continue) {
					validValues := []apiv1.TaskListKind{
						apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
						apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
						apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestPollForDecisionTaskResponse(t *testing.T) {
	for _, item := range []*apiv1.PollForDecisionTaskResponse{nil, {}, &testdata.PollForDecisionTaskResponse} {
		assert.Equal(t, item, proto.PollForDecisionTaskResponse(thrift.PollForDecisionTaskResponse(item)))
	}

	runFuzzTest(t,
		thrift.PollForDecisionTaskResponse,
		proto.PollForDecisionTaskResponse,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(p *apiv1.PollForDecisionTaskResponse, c fuzz.Continue) {
					p.TaskToken = make([]byte, c.Intn(50))
					c.Fuzz(&p.TaskToken)
					c.Fuzz(&p.StartedEventId)
					c.Fuzz(&p.Attempt)
					c.Fuzz(&p.BacklogCountHint)
					c.Fuzz(&p.NextEventId)
					c.Fuzz(&p.TotalHistoryBytes)
					c.Fuzz(&p.NextPageToken)
					c.Fuzz(&p.PreviousStartedEventId)
					c.Fuzz(&p.ScheduledTime)
					c.Fuzz(&p.StartedTime)
					c.Fuzz(&p.AutoConfigHint)
				},
			},
			ExcludedFields: []string{
				// [TOO HARD] These fields are causing issues when the entire struct is generated by gofuzz - even though
				// individual fields can be fuzzed successfully, and work in their isolated tests. They've been excluded implicitly
				// from the implementation of the CustomFunc by not implementing them, and are listed here for clarity.
				"WorkflowExecution",         // Complex nested WorkflowExecution struct - tested in TestWorkflowExecution
				"WorkflowType",              // Complex nested WorkflowType struct - tested in TestWorkflowType
				"History",                   // Complex nested History struct with HistoryEvent arrays - tested in TestHistory
				"Query",                     // Complex nested WorkflowQuery struct - tested in TestWorkflowQuery
				"WorkflowExecutionTaskList", // Complex nested TaskList struct - tested in TestTaskList
				"Queries",                   // Map of string to *WorkflowQuery - tested in TestWorkflowQueryMap
			},
		},
	)
}
func TestPollerInfo(t *testing.T) {
	for _, item := range []*apiv1.PollerInfo{nil, {}, &testdata.PollerInfo} {
		assert.Equal(t, item, proto.PollerInfo(thrift.PollerInfo(item)))
	}

	runFuzzTest(t,
		thrift.PollerInfo,
		proto.PollerInfo,
		FuzzOptions{},
	)
}
func TestQueryRejected(t *testing.T) {
	for _, item := range []*apiv1.QueryRejected{nil, {}, &testdata.QueryRejected} {
		assert.Equal(t, item, proto.QueryRejected(thrift.QueryRejected(item)))
	}

	runFuzzTest(t,
		thrift.QueryRejected,
		proto.QueryRejected,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.WorkflowExecutionCloseStatus, c fuzz.Continue) {
					validValues := []apiv1.WorkflowExecutionCloseStatus{
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_INVALID,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_COMPLETED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_FAILED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CANCELED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TERMINATED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CONTINUED_AS_NEW,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TIMED_OUT,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestQueryWorkflowRequest(t *testing.T) {
	for _, item := range []*apiv1.QueryWorkflowRequest{nil, {}, &testdata.QueryWorkflowRequest} {
		assert.Equal(t, item, proto.QueryWorkflowRequest(thrift.QueryWorkflowRequest(item)))
	}

	runFuzzTest(t,
		thrift.QueryWorkflowRequest,
		proto.QueryWorkflowRequest,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.QueryRejectCondition, c fuzz.Continue) {
					validValues := []apiv1.QueryRejectCondition{
						apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_INVALID,
						apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_NOT_OPEN,
						apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
				func(e *apiv1.QueryConsistencyLevel, c fuzz.Continue) {
					// Generate only valid QueryConsistencyLevel values
					validValues := []apiv1.QueryConsistencyLevel{
						apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_INVALID,
						apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_EVENTUAL,
						apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_STRONG,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestQueryWorkflowResponse(t *testing.T) {
	for _, item := range []*apiv1.QueryWorkflowResponse{nil, {}, &testdata.QueryWorkflowResponse} {
		assert.Equal(t, item, proto.QueryWorkflowResponse(thrift.QueryWorkflowResponse(item)))
	}

	runFuzzTest(t,
		thrift.QueryWorkflowResponse,
		proto.QueryWorkflowResponse,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.WorkflowExecutionCloseStatus, c fuzz.Continue) {
					validValues := []apiv1.WorkflowExecutionCloseStatus{
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_INVALID,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_COMPLETED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_FAILED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CANCELED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TERMINATED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CONTINUED_AS_NEW,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TIMED_OUT,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestRecordActivityTaskHeartbeatByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatByIDRequest{nil, {}, &testdata.RecordActivityTaskHeartbeatByIDRequest} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatByIDRequest(thrift.RecordActivityTaskHeartbeatByIDRequest(item)))
	}

	runFuzzTest(t,
		thrift.RecordActivityTaskHeartbeatByIDRequest,
		proto.RecordActivityTaskHeartbeatByIDRequest,
		FuzzOptions{},
	)
}
func TestRecordActivityTaskHeartbeatByIDResponse(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatByIDResponse{nil, {}, &testdata.RecordActivityTaskHeartbeatByIDResponse} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatByIDResponse(thrift.RecordActivityTaskHeartbeatByIDResponse(item)))
	}

	runFuzzTest(t,
		thrift.RecordActivityTaskHeartbeatByIDResponse,
		proto.RecordActivityTaskHeartbeatByIDResponse,
		FuzzOptions{},
	)
}
func TestRecordActivityTaskHeartbeatRequest(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatRequest{nil, {}, &testdata.RecordActivityTaskHeartbeatRequest} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatRequest(thrift.RecordActivityTaskHeartbeatRequest(item)))
	}

	runFuzzTest(t,
		thrift.RecordActivityTaskHeartbeatRequest,
		proto.RecordActivityTaskHeartbeatRequest,
		FuzzOptions{},
	)
}
func TestRecordActivityTaskHeartbeatResponse(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatResponse{nil, {}, &testdata.RecordActivityTaskHeartbeatResponse} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatResponse(thrift.RecordActivityTaskHeartbeatResponse(item)))
	}

	runFuzzTest(t,
		thrift.RecordActivityTaskHeartbeatResponse,
		proto.RecordActivityTaskHeartbeatResponse,
		FuzzOptions{},
	)
}
func TestRegisterDomainRequest(t *testing.T) {
	for _, item := range []*apiv1.RegisterDomainRequest{nil, &testdata.RegisterDomainRequest} {
		assert.Equal(t, item, proto.RegisterDomainRequest(thrift.RegisterDomainRequest(item)))
	}

	runFuzzTest(t,
		thrift.RegisterDomainRequest,
		proto.RegisterDomainRequest,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(status *apiv1.ArchivalStatus, c fuzz.Continue) {
					validValues := []apiv1.ArchivalStatus{
						apiv1.ArchivalStatus_ARCHIVAL_STATUS_INVALID,
						apiv1.ArchivalStatus_ARCHIVAL_STATUS_DISABLED,
						apiv1.ArchivalStatus_ARCHIVAL_STATUS_ENABLED,
					}
					*status = validValues[c.Intn(len(validValues))]
				},
				// Custom fuzzer for WorkflowExecutionRetentionPeriod - must be day-precision
				// because thrift mapping uses durationToDays (truncates to day boundaries)
				func(req *apiv1.RegisterDomainRequest, c fuzz.Continue) {
					if req.WorkflowExecutionRetentionPeriod != nil {
						// Generate days within int32 range: max ~5.8 million days (~16000 years)
						days := c.Int63n(MaxDurationSeconds / (24 * 3600))
						req.WorkflowExecutionRetentionPeriod.Seconds = days * 24 * 3600
						req.WorkflowExecutionRetentionPeriod.Nanos = 0
					}
				},
			},
		},
	)
}
func TestRequestCancelActivityTaskFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelActivityTaskFailedEventAttributes{nil, {}, &testdata.RequestCancelActivityTaskFailedEventAttributes} {
		assert.Equal(t, item, proto.RequestCancelActivityTaskFailedEventAttributes(thrift.RequestCancelActivityTaskFailedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.RequestCancelActivityTaskFailedEventAttributes,
		proto.RequestCancelActivityTaskFailedEventAttributes,
		FuzzOptions{},
	)
}
func TestRequestCancelExternalWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelExternalWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.RequestCancelExternalWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.RequestCancelExternalWorkflowExecutionFailedEventAttributes(thrift.RequestCancelExternalWorkflowExecutionFailedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.RequestCancelExternalWorkflowExecutionFailedEventAttributes,
		proto.RequestCancelExternalWorkflowExecutionFailedEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(cause *apiv1.CancelExternalWorkflowExecutionFailedCause, c fuzz.Continue) {
					validValues := []apiv1.CancelExternalWorkflowExecutionFailedCause{
						apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID,
						apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNKNOWN_EXTERNAL_WORKFLOW_EXECUTION,
					}
					*cause = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestRequestCancelExternalWorkflowExecutionInitiatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{nil, {}, &testdata.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes} {
		assert.Equal(t, item, proto.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes(thrift.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes,
		proto.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes,
		FuzzOptions{},
	)
}
func TestRequestCancelWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelWorkflowExecutionRequest{nil, {}, &testdata.RequestCancelWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.RequestCancelWorkflowExecutionRequest(thrift.RequestCancelWorkflowExecutionRequest(item)))
	}

	runFuzzTest(t,
		thrift.RequestCancelWorkflowExecutionRequest,
		proto.RequestCancelWorkflowExecutionRequest,
		FuzzOptions{
			ExcludedFields: []string{
				"Cause",               // [BUG] Cause is not mapped in thrift
				"FirstExecutionRunId", // [BUG] FirstExecutionRunId is not mapped in either mapper
			},
		},
	)
}
func TestResetPointInfo(t *testing.T) {
	for _, item := range []*apiv1.ResetPointInfo{nil, {}, &testdata.ResetPointInfo} {
		assert.Equal(t, item, proto.ResetPointInfo(thrift.ResetPointInfo(item)))
	}

	runFuzzTest(t,
		thrift.ResetPointInfo,
		proto.ResetPointInfo,
		FuzzOptions{},
	)
}
func TestResetPoints(t *testing.T) {
	for _, item := range []*apiv1.ResetPoints{nil, {}, &testdata.ResetPoints} {
		assert.Equal(t, item, proto.ResetPoints(thrift.ResetPoints(item)))
	}

	runFuzzTest(t,
		thrift.ResetPoints,
		proto.ResetPoints,
		FuzzOptions{},
	)
}
func TestResetStickyTaskListRequest(t *testing.T) {
	for _, item := range []*apiv1.ResetStickyTaskListRequest{nil, {}, &testdata.ResetStickyTaskListRequest} {
		assert.Equal(t, item, proto.ResetStickyTaskListRequest(thrift.ResetStickyTaskListRequest(item)))
	}

	runFuzzTest(t,
		thrift.ResetStickyTaskListRequest,
		proto.ResetStickyTaskListRequest,
		FuzzOptions{},
	)
}
func TestResetWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.ResetWorkflowExecutionRequest{nil, {}, &testdata.ResetWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.ResetWorkflowExecutionRequest(thrift.ResetWorkflowExecutionRequest(item)))
	}

	runFuzzTest(t,
		thrift.ResetWorkflowExecutionRequest,
		proto.ResetWorkflowExecutionRequest,
		FuzzOptions{},
	)
}
func TestResetWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.ResetWorkflowExecutionResponse{nil, {}, &testdata.ResetWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.ResetWorkflowExecutionResponse(thrift.ResetWorkflowExecutionResponse(item)))
	}

	runFuzzTest(t,
		thrift.ResetWorkflowExecutionResponse,
		proto.ResetWorkflowExecutionResponse,
		FuzzOptions{},
	)
}
func TestRespondActivityTaskCanceledByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCanceledByIDRequest{nil, {}, &testdata.RespondActivityTaskCanceledByIDRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCanceledByIDRequest(thrift.RespondActivityTaskCanceledByIDRequest(item)))
	}

	runFuzzTest(t,
		thrift.RespondActivityTaskCanceledByIDRequest,
		proto.RespondActivityTaskCanceledByIDRequest,
		FuzzOptions{},
	)
}
func TestRespondActivityTaskCanceledRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCanceledRequest{nil, {}, &testdata.RespondActivityTaskCanceledRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCanceledRequest(thrift.RespondActivityTaskCanceledRequest(item)))
	}

	runFuzzTest(t,
		thrift.RespondActivityTaskCanceledRequest,
		proto.RespondActivityTaskCanceledRequest,
		FuzzOptions{},
	)
}
func TestRespondActivityTaskCompletedByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCompletedByIDRequest{nil, {}, &testdata.RespondActivityTaskCompletedByIDRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCompletedByIDRequest(thrift.RespondActivityTaskCompletedByIDRequest(item)))
	}

	runFuzzTest(t,
		thrift.RespondActivityTaskCompletedByIDRequest,
		proto.RespondActivityTaskCompletedByIDRequest,
		FuzzOptions{},
	)
}
func TestRespondActivityTaskCompletedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCompletedRequest{nil, {}, &testdata.RespondActivityTaskCompletedRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCompletedRequest(thrift.RespondActivityTaskCompletedRequest(item)))
	}

	runFuzzTest(t,
		thrift.RespondActivityTaskCompletedRequest,
		proto.RespondActivityTaskCompletedRequest,
		FuzzOptions{},
	)
}
func TestRespondActivityTaskFailedByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskFailedByIDRequest{nil, {}, &testdata.RespondActivityTaskFailedByIDRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskFailedByIDRequest(thrift.RespondActivityTaskFailedByIDRequest(item)))
	}

	runFuzzTest(t,
		thrift.RespondActivityTaskFailedByIDRequest,
		proto.RespondActivityTaskFailedByIDRequest,
		FuzzOptions{},
	)
}
func TestRespondActivityTaskFailedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskFailedRequest{nil, {}, &testdata.RespondActivityTaskFailedRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskFailedRequest(thrift.RespondActivityTaskFailedRequest(item)))
	}

	runFuzzTest(t,
		thrift.RespondActivityTaskFailedRequest,
		proto.RespondActivityTaskFailedRequest,
		FuzzOptions{},
	)
}
func TestRespondDecisionTaskCompletedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondDecisionTaskCompletedRequest{nil, {}, &testdata.RespondDecisionTaskCompletedRequest} {
		assert.Equal(t, item, proto.RespondDecisionTaskCompletedRequest(thrift.RespondDecisionTaskCompletedRequest(item)))
	}

	runFuzzTest(t,
		thrift.RespondDecisionTaskCompletedRequest,
		proto.RespondDecisionTaskCompletedRequest,
		FuzzOptions{
			NilChance: DefaultNilChance,
			CustomFuncs: []interface{}{
				func(req *apiv1.RespondDecisionTaskCompletedRequest, c fuzz.Continue) {
					c.Fuzz(&req.TaskToken)
					c.Fuzz(&req.ExecutionContext)
					c.Fuzz(&req.Identity)
					c.Fuzz(&req.ReturnNewDecisionTask)
					c.Fuzz(&req.ForceCreateNewDecisionTask)
					c.Fuzz(&req.BinaryChecksum)
				},
			},
			ExcludedFields: []string{
				"Decisions",        // [TOO HARD] Array of complex Decision structures - tested in TestDecisionArray
				"StickyAttributes", // [TOO HARD] Complex nested StickyExecutionAttributes structure - tested in TestStickyExecutionAttributes
				"QueryResults",     // [TOO HARD] Map of complex WorkflowQueryResult structures - tested in TestWorkflowQueryResultMap
			},
		},
	)
}
func TestRespondDecisionTaskCompletedResponse(t *testing.T) {
	for _, item := range []*apiv1.RespondDecisionTaskCompletedResponse{nil, {}, &testdata.RespondDecisionTaskCompletedResponse} {
		assert.Equal(t, item, proto.RespondDecisionTaskCompletedResponse(thrift.RespondDecisionTaskCompletedResponse(item)))
	}

	runFuzzTest(t,
		thrift.RespondDecisionTaskCompletedResponse,
		proto.RespondDecisionTaskCompletedResponse,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(resp *apiv1.RespondDecisionTaskCompletedResponse, c fuzz.Continue) {
					resp.DecisionTask = &apiv1.PollForDecisionTaskResponse{}
					c.Fuzz(&resp.DecisionTask.TaskToken)
					c.Fuzz(&resp.DecisionTask.StartedEventId)
					c.Fuzz(&resp.DecisionTask.Attempt)
				},
			},
			ExcludedFields: []string{
				"ActivitiesToDispatchLocally", // [TOO HARD] Map of complex types that causes protobuf field issues
			},
		},
	)
}
func TestRespondDecisionTaskFailedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondDecisionTaskFailedRequest{nil, {}, &testdata.RespondDecisionTaskFailedRequest} {
		assert.Equal(t, item, proto.RespondDecisionTaskFailedRequest(thrift.RespondDecisionTaskFailedRequest(item)))
	}

	runFuzzTest(t,
		thrift.RespondDecisionTaskFailedRequest,
		proto.RespondDecisionTaskFailedRequest,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(cause *apiv1.DecisionTaskFailedCause, c fuzz.Continue) {
					validValues := []apiv1.DecisionTaskFailedCause{
						apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_UNHANDLED_DECISION,
						apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SCHEDULE_ACTIVITY_ATTRIBUTES,
						apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_ACTIVITY_ATTRIBUTES,
						apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_START_TIMER_ATTRIBUTES,
						apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_TIMER_ATTRIBUTES,
					}
					*cause = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestRespondQueryTaskCompletedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondQueryTaskCompletedRequest{nil, {Result: &apiv1.WorkflowQueryResult{}}, &testdata.RespondQueryTaskCompletedRequest} {
		assert.Equal(t, item, proto.RespondQueryTaskCompletedRequest(thrift.RespondQueryTaskCompletedRequest(item)))
	}

	runFuzzTest(t,
		thrift.RespondQueryTaskCompletedRequest,
		proto.RespondQueryTaskCompletedRequest,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(resultType *apiv1.QueryResultType, c fuzz.Continue) {
					validValues := []apiv1.QueryResultType{
						apiv1.QueryResultType_QUERY_RESULT_TYPE_INVALID,
						apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED,
						apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED,
					}
					*resultType = validValues[c.Intn(len(validValues))]
				},
			},
			ExcludedFields: []string{
				"Result", // [BUG] Result nil values being converted to non-nil objects
			},
		},
	)
}
func TestRetryPolicy(t *testing.T) {
	for _, item := range []*apiv1.RetryPolicy{nil, {}, &testdata.RetryPolicy} {
		assert.Equal(t, item, proto.RetryPolicy(thrift.RetryPolicy(item)))
	}

	runFuzzTest(t,
		thrift.RetryPolicy,
		proto.RetryPolicy,
		FuzzOptions{
			ExcludedFields: []string{},
		},
	)
}
func TestScanWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ScanWorkflowExecutionsRequest{nil, {}, &testdata.ScanWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.ScanWorkflowExecutionsRequest(thrift.ScanWorkflowExecutionsRequest(item)))
	}

	runFuzzTest(t,
		thrift.ScanWorkflowExecutionsRequest,
		proto.ScanWorkflowExecutionsRequest,
		FuzzOptions{},
	)
}
func TestScanWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ScanWorkflowExecutionsResponse{nil, {}, &testdata.ScanWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ScanWorkflowExecutionsResponse(thrift.ScanWorkflowExecutionsResponse(item)))
	}

	runFuzzTest(t,
		thrift.ScanWorkflowExecutionsResponse,
		proto.ScanWorkflowExecutionsResponse,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(resp *apiv1.ScanWorkflowExecutionsResponse, c fuzz.Continue) {
					c.Fuzz(&resp.NextPageToken)
				},
			},
			ExcludedFields: []string{
				"Executions", // [TOO HARD] Array of complex WorkflowExecutionInfo structures - tested in TestWorkflowExecutionInfo
			},
		},
	)
}
func TestSearchAttributes(t *testing.T) {
	for _, item := range []*apiv1.SearchAttributes{nil, {}, &testdata.SearchAttributes} {
		assert.Equal(t, item, proto.SearchAttributes(thrift.SearchAttributes(item)))
	}

	runFuzzTest(t,
		thrift.SearchAttributes,
		proto.SearchAttributes,
		FuzzOptions{},
	)
}
func TestSignalExternalWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.SignalExternalWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.SignalExternalWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.SignalExternalWorkflowExecutionFailedEventAttributes(thrift.SignalExternalWorkflowExecutionFailedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.SignalExternalWorkflowExecutionFailedEventAttributes,
		proto.SignalExternalWorkflowExecutionFailedEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(cause *apiv1.SignalExternalWorkflowExecutionFailedCause, c fuzz.Continue) {
					validValues := []apiv1.SignalExternalWorkflowExecutionFailedCause{
						apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID,
						apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNKNOWN_EXTERNAL_WORKFLOW_EXECUTION,
						apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_COMPLETED,
					}
					*cause = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestSignalExternalWorkflowExecutionInitiatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.SignalExternalWorkflowExecutionInitiatedEventAttributes{nil, {}, &testdata.SignalExternalWorkflowExecutionInitiatedEventAttributes} {
		assert.Equal(t, item, proto.SignalExternalWorkflowExecutionInitiatedEventAttributes(thrift.SignalExternalWorkflowExecutionInitiatedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.SignalExternalWorkflowExecutionInitiatedEventAttributes,
		proto.SignalExternalWorkflowExecutionInitiatedEventAttributes,
		FuzzOptions{},
	)
}
func TestSignalWithStartWorkflowExecutionRequest(t *testing.T) {
	tests := []*apiv1.SignalWithStartWorkflowExecutionRequest{
		nil,
		{StartRequest: &apiv1.StartWorkflowExecutionRequest{}},
		&testdata.SignalWithStartWorkflowExecutionRequest,
		&testdata.SignalWithStartWorkflowExecutionRequestWithCronAndActiveClusterSelectionPolicy1,
		&testdata.SignalWithStartWorkflowExecutionRequestWithCronAndActiveClusterSelectionPolicy2,
	}
	for i, item := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			assert.Equal(t, item, proto.SignalWithStartWorkflowExecutionRequest(thrift.SignalWithStartWorkflowExecutionRequest(item)))
		})
	}

	runFuzzTest(t,
		thrift.SignalWithStartWorkflowExecutionRequest,
		proto.SignalWithStartWorkflowExecutionRequest,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TaskListKind, c fuzz.Continue) {
					validValues := []apiv1.TaskListKind{
						apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
						apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
						apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
				func(req *apiv1.SignalWithStartWorkflowExecutionRequest, c fuzz.Continue) {
					c.Fuzz(&req.SignalName)
					c.Fuzz(&req.Control)
					// Safely create a minimal StartRequest to avoid nil issues
					req.StartRequest = &apiv1.StartWorkflowExecutionRequest{
						Domain:     c.RandString(),
						WorkflowId: c.RandString(),
					}
				},
			},
			ExcludedFields: []string{
				"SignalInput", // [TOO HARD] Complex Payload structure that causes gofuzz nil panics
			},
		},
	)
}
func TestSignalWithStartWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.SignalWithStartWorkflowExecutionResponse{nil, {}, &testdata.SignalWithStartWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.SignalWithStartWorkflowExecutionResponse(thrift.SignalWithStartWorkflowExecutionResponse(item)))
	}

	runFuzzTest(t,
		thrift.SignalWithStartWorkflowExecutionResponse,
		proto.SignalWithStartWorkflowExecutionResponse,
		FuzzOptions{},
	)
}
func TestSignalWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.SignalWorkflowExecutionRequest{nil, {}, &testdata.SignalWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.SignalWorkflowExecutionRequest(thrift.SignalWorkflowExecutionRequest(item)))
	}

	runFuzzTest(t,
		thrift.SignalWorkflowExecutionRequest,
		proto.SignalWorkflowExecutionRequest,
		FuzzOptions{},
	)
}
func TestStartChildWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.StartChildWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.StartChildWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.StartChildWorkflowExecutionFailedEventAttributes(thrift.StartChildWorkflowExecutionFailedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.StartChildWorkflowExecutionFailedEventAttributes,
		proto.StartChildWorkflowExecutionFailedEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(cause *apiv1.ChildWorkflowExecutionFailedCause, c fuzz.Continue) {
					validValues := []apiv1.ChildWorkflowExecutionFailedCause{
						apiv1.ChildWorkflowExecutionFailedCause_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID,
						apiv1.ChildWorkflowExecutionFailedCause_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_RUNNING,
					}
					*cause = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestStartChildWorkflowExecutionInitiatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.StartChildWorkflowExecutionInitiatedEventAttributes{nil, {}, &testdata.StartChildWorkflowExecutionInitiatedEventAttributes} {
		assert.Equal(t, item, proto.StartChildWorkflowExecutionInitiatedEventAttributes(thrift.StartChildWorkflowExecutionInitiatedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.StartChildWorkflowExecutionInitiatedEventAttributes,
		proto.StartChildWorkflowExecutionInitiatedEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TaskListKind, c fuzz.Continue) {
					validValues := []apiv1.TaskListKind{
						apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
						apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
						apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
				// The rest of the attributes are fuzzed here, as gofuzz panics when the entire struct needs to be fuzzed
				// But it works fine when each individual field is fuzzed indepdenently
				func(attr *apiv1.StartChildWorkflowExecutionInitiatedEventAttributes, c fuzz.Continue) {
					c.Fuzz(&attr.Domain)
					c.Fuzz(&attr.WorkflowId)
					c.Fuzz(&attr.Control)
					c.Fuzz(&attr.DecisionTaskCompletedEventId)
					c.Fuzz(&attr.CronSchedule)
					c.Fuzz(&attr.TaskList)
					c.Fuzz(&attr.WorkflowType)
					c.Fuzz(&attr.RetryPolicy)
					c.Fuzz(&attr.Header)
				},
			},
		},
	)
}
func TestStartTimeFilter(t *testing.T) {
	for _, item := range []*apiv1.StartTimeFilter{nil, {}, &testdata.StartTimeFilter} {
		assert.Equal(t, item, proto.StartTimeFilter(thrift.StartTimeFilter(item)))
	}

	runFuzzTest(t,
		thrift.StartTimeFilter,
		proto.StartTimeFilter,
		FuzzOptions{},
	)
}
func TestStartWorkflowExecutionRequest(t *testing.T) {
	tests := []*apiv1.StartWorkflowExecutionRequest{
		nil,
		{},
		&testdata.StartWorkflowExecutionRequest,
		&testdata.StartWorkflowExecutionRequestWithCronAndActiveClusterSelectionPolicy1,
		&testdata.StartWorkflowExecutionRequestWithCronAndActiveClusterSelectionPolicy2,
	}
	for i, item := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			assert.Equal(t, item, proto.StartWorkflowExecutionRequest(thrift.StartWorkflowExecutionRequest(item)))
		})
	}

	runFuzzTest(t,
		thrift.StartWorkflowExecutionRequest,
		proto.StartWorkflowExecutionRequest,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TaskListKind, c fuzz.Continue) {
					validValues := []apiv1.TaskListKind{
						apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
						apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
						apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
				func(policy *apiv1.WorkflowIdReusePolicy, c fuzz.Continue) {
					validValues := []apiv1.WorkflowIdReusePolicy{
						apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_INVALID,
						apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
						apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
						apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
						apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
					}
					*policy = validValues[c.Intn(len(validValues))]
				},
				func(policy *apiv1.CronOverlapPolicy, c fuzz.Continue) {
					validValues := []apiv1.CronOverlapPolicy{
						apiv1.CronOverlapPolicy_CRON_OVERLAP_POLICY_INVALID,
						apiv1.CronOverlapPolicy_CRON_OVERLAP_POLICY_SKIPPED,
						apiv1.CronOverlapPolicy_CRON_OVERLAP_POLICY_BUFFER_ONE,
					}
					*policy = validValues[c.Intn(len(validValues))]
				},
				// Fuzz the entire request as gofuzz panics when it tries to populate the entire complex object.
				// Populating the individual fields independently works though.
				func(req *apiv1.StartWorkflowExecutionRequest, c fuzz.Continue) {
					c.Fuzz(&req.Domain)
					c.Fuzz(&req.WorkflowId)
					c.Fuzz(&req.WorkflowType)
					c.Fuzz(&req.TaskList)
					c.Fuzz(&req.Input)
					c.Fuzz(&req.ExecutionStartToCloseTimeout)
					c.Fuzz(&req.TaskStartToCloseTimeout)
					c.Fuzz(&req.Identity)
					c.Fuzz(&req.RequestId)
					c.Fuzz(&req.WorkflowIdReusePolicy)
					c.Fuzz(&req.RetryPolicy)
					c.Fuzz(&req.CronSchedule)
					c.Fuzz(&req.Memo)
					c.Fuzz(&req.SearchAttributes)
					c.Fuzz(&req.Header)
					c.Fuzz(&req.DelayStart)
					c.Fuzz(&req.JitterStart)
					c.Fuzz(&req.FirstRunAt)
					c.Fuzz(&req.CronOverlapPolicy)
				},
			},
			ExcludedFields: []string{
				// [BUG] ActiveClusterSelectionPolicy appears to be failing the round trip
				// TODO: Investigate the mappers and determine where it is failing fuzz testing
				"ActiveClusterSelectionPolicy",
			},
		},
	)
}
func TestStartWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.StartWorkflowExecutionResponse{nil, {}, &testdata.StartWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.StartWorkflowExecutionResponse(thrift.StartWorkflowExecutionResponse(item)))
	}

	runFuzzTest(t,
		thrift.StartWorkflowExecutionResponse,
		proto.StartWorkflowExecutionResponse,
		FuzzOptions{},
	)
}
func TestStatusFilter(t *testing.T) {
	for _, item := range []*apiv1.StatusFilter{nil, &testdata.StatusFilter} {
		assert.Equal(t, item, proto.StatusFilter(thrift.StatusFilter(item)))
	}

	runFuzzTest(t,
		thrift.StatusFilter,
		proto.StatusFilter,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.WorkflowExecutionCloseStatus, c fuzz.Continue) {
					// TODO: WORKFLOW_EXECUTION_CLOSE_STATUS_INVALID panics as an unexpected value
					validValues := []apiv1.WorkflowExecutionCloseStatus{
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_COMPLETED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_FAILED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CANCELED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TERMINATED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CONTINUED_AS_NEW,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TIMED_OUT,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestStickyExecutionAttributes(t *testing.T) {
	for _, item := range []*apiv1.StickyExecutionAttributes{nil, {}, &testdata.StickyExecutionAttributes} {
		assert.Equal(t, item, proto.StickyExecutionAttributes(thrift.StickyExecutionAttributes(item)))
	}

	runFuzzTest(t,
		thrift.StickyExecutionAttributes,
		proto.StickyExecutionAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TaskListKind, c fuzz.Continue) {
					validValues := []apiv1.TaskListKind{
						apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
						apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
						apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
			ExcludedFields: []string{},
		},
	)
}
func TestSupportedClientVersions(t *testing.T) {
	for _, item := range []*apiv1.SupportedClientVersions{nil, {}, &testdata.SupportedClientVersions} {
		assert.Equal(t, item, proto.SupportedClientVersions(thrift.SupportedClientVersions(item)))
	}

	runFuzzTest(t,
		thrift.SupportedClientVersions,
		proto.SupportedClientVersions,
		FuzzOptions{},
	)
}
func TestTaskIDBlock(t *testing.T) {
	for _, item := range []*apiv1.TaskIDBlock{nil, {}, &testdata.TaskIDBlock} {
		assert.Equal(t, item, proto.TaskIDBlock(thrift.TaskIDBlock(item)))
	}

	runFuzzTest(t,
		thrift.TaskIDBlock,
		proto.TaskIDBlock,
		FuzzOptions{},
	)
}
func TestTaskList(t *testing.T) {
	for _, item := range []*apiv1.TaskList{nil, {}, &testdata.TaskList} {
		assert.Equal(t, item, proto.TaskList(thrift.TaskList(item)))
	}

	runFuzzTest(t,
		thrift.TaskList,
		proto.TaskList,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TaskListKind, c fuzz.Continue) {
					validValues := []apiv1.TaskListKind{
						apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
						apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
						apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}

func TestTaskListMetadata(t *testing.T) {
	for _, item := range []*apiv1.TaskListMetadata{nil, {}, &testdata.TaskListMetadata} {
		assert.Equal(t, item, proto.TaskListMetadata(thrift.TaskListMetadata(item)))
	}

	runFuzzTest(t,
		thrift.TaskListMetadata,
		proto.TaskListMetadata,
		FuzzOptions{},
	)
}

func TestTaskListPartitionMetadata(t *testing.T) {
	for _, item := range []*apiv1.TaskListPartitionMetadata{nil, {}, &testdata.TaskListPartitionMetadata} {
		assert.Equal(t, item, proto.TaskListPartitionMetadata(thrift.TaskListPartitionMetadata(item)))
	}

	runFuzzTest(t,
		thrift.TaskListPartitionMetadata,
		proto.TaskListPartitionMetadata,
		FuzzOptions{},
	)
}
func TestTaskListStatus(t *testing.T) {
	for _, item := range []*apiv1.TaskListStatus{nil, {}, &testdata.TaskListStatus} {
		assert.Equal(t, item, proto.TaskListStatus(thrift.TaskListStatus(item)))
	}

	runFuzzTest(t,
		thrift.TaskListStatus,
		proto.TaskListStatus,
		FuzzOptions{
			ExcludedFields: []string{
				"NewTasksPerSecond",     // [BUG] NewTasksPerSecond is not mapped
				"IsolationGroupMetrics", // [BUG] IsolationGroupMetrics is not mapped
			},
		},
	)
}
func TestTerminateWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.TerminateWorkflowExecutionRequest{nil, {}, &testdata.TerminateWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.TerminateWorkflowExecutionRequest(thrift.TerminateWorkflowExecutionRequest(item)))
	}

	runFuzzTest(t,
		thrift.TerminateWorkflowExecutionRequest,
		proto.TerminateWorkflowExecutionRequest,
		FuzzOptions{
			ExcludedFields: []string{
				"FirstExecutionRunId", // [BUG] FirstExecutionRunId is not mapped
			},
		},
	)
}
func TestTimerCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.TimerCanceledEventAttributes{nil, {}, &testdata.TimerCanceledEventAttributes} {
		assert.Equal(t, item, proto.TimerCanceledEventAttributes(thrift.TimerCanceledEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.TimerCanceledEventAttributes,
		proto.TimerCanceledEventAttributes,
		FuzzOptions{},
	)
}
func TestTimerFiredEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.TimerFiredEventAttributes{nil, {}, &testdata.TimerFiredEventAttributes} {
		assert.Equal(t, item, proto.TimerFiredEventAttributes(thrift.TimerFiredEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.TimerFiredEventAttributes,
		proto.TimerFiredEventAttributes,
		FuzzOptions{},
	)
}
func TestTimerStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.TimerStartedEventAttributes{nil, {}, &testdata.TimerStartedEventAttributes} {
		assert.Equal(t, item, proto.TimerStartedEventAttributes(thrift.TimerStartedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.TimerStartedEventAttributes,
		proto.TimerStartedEventAttributes,
		FuzzOptions{},
	)
}
func TestUpdateDomainRequest(t *testing.T) {
	for _, item := range []*apiv1.UpdateDomainRequest{nil, {UpdateMask: &gogo.FieldMask{}}, &testdata.UpdateDomainRequest} {
		assert.Equal(t, item, proto.UpdateDomainRequest(thrift.UpdateDomainRequest(item)))
	}

	runFuzzTest(t,
		thrift.UpdateDomainRequest,
		proto.UpdateDomainRequest,
		FuzzOptions{
			CustomFuncs: []interface{}{
				// [INVALID DATA] ArchivalStatus enum values - gofuzz generates invalid enum values that cause "unexpected enum value" panics
				func(status *apiv1.ArchivalStatus, c fuzz.Continue) {
					validValues := []apiv1.ArchivalStatus{
						apiv1.ArchivalStatus_ARCHIVAL_STATUS_INVALID,
						apiv1.ArchivalStatus_ARCHIVAL_STATUS_DISABLED,
						apiv1.ArchivalStatus_ARCHIVAL_STATUS_ENABLED,
					}
					*status = validValues[c.Intn(len(validValues))]
				},
				// Custom fuzzer to handle all safe fields properly
				func(req *apiv1.UpdateDomainRequest, c fuzz.Continue) {
					// Fuzz basic fields that are safe to fuzz
					c.Fuzz(&req.SecurityToken)
					c.Fuzz(&req.Name)
					c.Fuzz(&req.Description)
					c.Fuzz(&req.OwnerEmail)
					c.Fuzz(&req.Data)
					c.Fuzz(&req.HistoryArchivalStatus)
					c.Fuzz(&req.HistoryArchivalUri)
					c.Fuzz(&req.VisibilityArchivalStatus)
					c.Fuzz(&req.VisibilityArchivalUri)
					c.Fuzz(&req.ActiveClusterName)
					c.Fuzz(&req.DeleteBadBinary)
					c.Fuzz(&req.FailoverTimeout)

					// Custom fuzzer for WorkflowExecutionRetentionPeriod - must be day-precision
					// because thrift mapping uses durationToDays (truncates to day boundaries)
					if req.WorkflowExecutionRetentionPeriod != nil {
						days := c.Int63n(MaxDurationSeconds / (24 * 3600))
						req.WorkflowExecutionRetentionPeriod.Seconds = days * 24 * 3600
						req.WorkflowExecutionRetentionPeriod.Nanos = 0
					}
				},
			},
			ExcludedFields: []string{
				"UpdateMask",     // [TOO HARD] Complex nested structure with protobuf metadata issues - mapper incorrectly populates UpdateMask paths
				"BadBinaries",    // [TOO HARD] Appears to be a fuzzing issue, tested in TestBadBinaries
				"Clusters",       // [TOO HARD] Appears to be a fuzzing issue
				"ActiveClusters", // [TOO HARD] Appears to be a fuzzing issue
			},
		},
	)
}
func TestUpdateDomainResponse(t *testing.T) {
	for _, item := range []*apiv1.UpdateDomainResponse{nil, &testdata.UpdateDomainResponse} {
		assert.Equal(t, item, proto.UpdateDomainResponse(thrift.UpdateDomainResponse(item)))
	}

	runFuzzTest(t,
		thrift.UpdateDomainResponse,
		proto.UpdateDomainResponse,
		FuzzOptions{
			// TODO: Re-enable NilChance and fix the mapper
			NilChance: 0.0,
			CustomFuncs: []interface{}{
				func(resp *apiv1.UpdateDomainResponse, c fuzz.Continue) {
					// Always create a valid Domain to avoid mapper nil-return issue
					resp.Domain = &apiv1.Domain{
						Name:                     c.RandString(),
						Status:                   apiv1.DomainStatus_DOMAIN_STATUS_REGISTERED,
						HistoryArchivalStatus:    apiv1.ArchivalStatus_ARCHIVAL_STATUS_DISABLED,
						VisibilityArchivalStatus: apiv1.ArchivalStatus_ARCHIVAL_STATUS_DISABLED,
					}
					// Set WorkflowExecutionRetentionPeriod with day-precision to avoid truncation
					if c.RandBool() {
						days := c.Int63n(MaxDurationSeconds / (24 * 3600))
						resp.Domain.WorkflowExecutionRetentionPeriod = &gogo.Duration{
							Seconds: days * 24 * 3600,
							Nanos:   0,
						}
					}
				},
			},
			ExcludedFields: []string{
				// Exclude nested fields that have complex issues like in DescribeDomainResponse
				"Domain.ActiveClusters",      // Nil pointer dereference in mapper conversion
				"Domain.Clusters",            // Protobuf metadata issues in nested ClusterReplicationConfiguration
				"Domain.FailoverInfo",        // Protobuf metadata issues in nested structures
				"Domain.IsolationGroups",     // Protobuf metadata issues in nested structures
				"Domain.AsyncWorkflowConfig", // Protobuf metadata issues in nested structures
				"Domain.BadBinaries",         // Protobuf metadata issues in nested structures
			},
		},
	)
}
func TestUpsertWorkflowSearchAttributesEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.UpsertWorkflowSearchAttributesEventAttributes{nil, {}, &testdata.UpsertWorkflowSearchAttributesEventAttributes} {
		assert.Equal(t, item, proto.UpsertWorkflowSearchAttributesEventAttributes(thrift.UpsertWorkflowSearchAttributesEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.UpsertWorkflowSearchAttributesEventAttributes,
		proto.UpsertWorkflowSearchAttributesEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkerVersionInfo(t *testing.T) {
	for _, item := range []*apiv1.WorkerVersionInfo{nil, {}, &testdata.WorkerVersionInfo} {
		assert.Equal(t, item, proto.WorkerVersionInfo(thrift.WorkerVersionInfo(item)))
	}

	runFuzzTest(t,
		thrift.WorkerVersionInfo,
		proto.WorkerVersionInfo,
		FuzzOptions{},
	)
}
func TestWorkflowExecution(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecution{nil, {}, &testdata.WorkflowExecution} {
		assert.Equal(t, item, proto.WorkflowExecution(thrift.WorkflowExecution(item)))
	}
	assert.Empty(t, thrift.WorkflowID(nil))
	assert.Empty(t, thrift.RunID(nil))

	runFuzzTest(t,
		thrift.WorkflowExecution,
		proto.WorkflowExecution,
		FuzzOptions{},
	)
}
func TestExternalExecutionInfo(t *testing.T) {
	assert.Nil(t, proto.ExternalExecutionInfo(nil, nil))
	assert.Nil(t, thrift.ExternalWorkflowExecution(nil))
	assert.Nil(t, thrift.ExternalInitiatedID(nil))
	assert.Panics(t, func() { proto.ExternalExecutionInfo(nil, common.Int64Ptr(testdata.EventID1)) })
	assert.Panics(t, func() { proto.ExternalExecutionInfo(thrift.WorkflowExecution(&testdata.WorkflowExecution), nil) })
	info := proto.ExternalExecutionInfo(thrift.WorkflowExecution(&testdata.WorkflowExecution), common.Int64Ptr(testdata.EventID1))
	assert.Equal(t, testdata.WorkflowExecution, *info.WorkflowExecution)
	assert.Equal(t, testdata.EventID1, info.InitiatedId)
}
func TestWorkflowExecutionCancelRequestedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionCancelRequestedEventAttributes{nil, {}, &testdata.WorkflowExecutionCancelRequestedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionCancelRequestedEventAttributes(thrift.WorkflowExecutionCancelRequestedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowExecutionCancelRequestedEventAttributes,
		proto.WorkflowExecutionCancelRequestedEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(attr *apiv1.WorkflowExecutionCancelRequestedEventAttributes, c fuzz.Continue) {
					// Handle ExternalExecutionInfo validation - either all fields set or none
					if c.RandBool() {
						// Set all fields for valid ExternalExecutionInfo
						attr.ExternalExecutionInfo = &apiv1.ExternalExecutionInfo{
							WorkflowExecution: &apiv1.WorkflowExecution{
								WorkflowId: c.RandString(),
								RunId:      c.RandString(),
							},
							InitiatedId: c.Int63(),
						}
					} else {
						// Set no external execution info
						attr.ExternalExecutionInfo = nil
					}
					attr.Cause = c.RandString()
					attr.Identity = c.RandString()
				},
			},
			ExcludedFields: []string{
				"RequestId", // TODO: RequestId is not mapped
			},
		},
	)
}
func TestWorkflowExecutionCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionCanceledEventAttributes{nil, {}, &testdata.WorkflowExecutionCanceledEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionCanceledEventAttributes(thrift.WorkflowExecutionCanceledEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowExecutionCanceledEventAttributes,
		proto.WorkflowExecutionCanceledEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkflowExecutionCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionCompletedEventAttributes{nil, {}, &testdata.WorkflowExecutionCompletedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionCompletedEventAttributes(thrift.WorkflowExecutionCompletedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowExecutionCompletedEventAttributes,
		proto.WorkflowExecutionCompletedEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkflowExecutionConfiguration(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionConfiguration{nil, {}, &testdata.WorkflowExecutionConfiguration} {
		assert.Equal(t, item, proto.WorkflowExecutionConfiguration(thrift.WorkflowExecutionConfiguration(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowExecutionConfiguration,
		proto.WorkflowExecutionConfiguration,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TaskListKind, c fuzz.Continue) {
					validValues := []apiv1.TaskListKind{
						apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
						apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
						apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
			ExcludedFields: []string{},
		},
	)
}
func TestWorkflowExecutionContinuedAsNewEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionContinuedAsNewEventAttributes{nil, {}, &testdata.WorkflowExecutionContinuedAsNewEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionContinuedAsNewEventAttributes(thrift.WorkflowExecutionContinuedAsNewEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowExecutionContinuedAsNewEventAttributes,
		proto.WorkflowExecutionContinuedAsNewEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TaskListKind, c fuzz.Continue) {
					// Generate only valid TaskListKind values
					validValues := []apiv1.TaskListKind{
						apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
						apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
						apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
				func(e *apiv1.ContinueAsNewInitiator, c fuzz.Continue) {
					// Generate only valid ContinueAsNewInitiator values
					validValues := []apiv1.ContinueAsNewInitiator{
						apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_INVALID,
						apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_DECIDER,
						apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_RETRY_POLICY,
						apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_CRON_SCHEDULE,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
			ExcludedFields: []string{},
		},
	)
}
func TestWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionFailedEventAttributes{nil, {}, &testdata.WorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionFailedEventAttributes(thrift.WorkflowExecutionFailedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowExecutionFailedEventAttributes,
		proto.WorkflowExecutionFailedEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkflowExecutionFilter(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionFilter{nil, {}, &testdata.WorkflowExecutionFilter} {
		assert.Equal(t, item, proto.WorkflowExecutionFilter(thrift.WorkflowExecutionFilter(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowExecutionFilter,
		proto.WorkflowExecutionFilter,
		FuzzOptions{},
	)
}
func TestParentExecutionInfo(t *testing.T) {
	assert.Nil(t, proto.ParentExecutionInfo(nil, nil, nil, nil))
	assert.Panics(t, func() { proto.ParentExecutionInfo(nil, &testdata.ParentExecutionInfo.DomainName, nil, nil) })
	info := proto.ParentExecutionInfo(nil,
		&testdata.ParentExecutionInfo.DomainName,
		thrift.WorkflowExecution(testdata.ParentExecutionInfo.WorkflowExecution),
		&testdata.ParentExecutionInfo.InitiatedId)
	assert.Equal(t, "", info.DomainId)
	assert.Equal(t, testdata.ParentExecutionInfo.DomainName, info.DomainName)
	assert.Equal(t, testdata.ParentExecutionInfo.WorkflowExecution, info.WorkflowExecution)
	assert.Equal(t, testdata.ParentExecutionInfo.InitiatedId, info.InitiatedId)
}

func TestWorkflowExecutionInfo(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionInfo{nil, {}, &testdata.WorkflowExecutionInfo} {
		assert.Equal(t, item, proto.WorkflowExecutionInfo(thrift.WorkflowExecutionInfo(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowExecutionInfo,
		proto.WorkflowExecutionInfo,
		FuzzOptions{
			NilChance: 0.0,
			CustomFuncs: []interface{}{
				func(e *apiv1.WorkflowExecutionCloseStatus, c fuzz.Continue) {
					validValues := []apiv1.WorkflowExecutionCloseStatus{
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_INVALID,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_COMPLETED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_FAILED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CANCELED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TERMINATED,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CONTINUED_AS_NEW,
						apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TIMED_OUT,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
				func(e *apiv1.CronOverlapPolicy, c fuzz.Continue) {
					validValues := []apiv1.CronOverlapPolicy{
						apiv1.CronOverlapPolicy_CRON_OVERLAP_POLICY_INVALID,
						apiv1.CronOverlapPolicy_CRON_OVERLAP_POLICY_SKIPPED,
						apiv1.CronOverlapPolicy_CRON_OVERLAP_POLICY_BUFFER_ONE,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
				func(e *apiv1.TaskListKind, c fuzz.Continue) {
					validValues := []apiv1.TaskListKind{
						apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
						apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
						apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
						apiv1.TaskListKind_TASK_LIST_KIND_EPHEMERAL,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
				// Custom fuzzer for WorkflowExecutionInfo that uses c.Fuzz for maximum coverage
				func(info *apiv1.WorkflowExecutionInfo, c fuzz.Continue) {
					c.Fuzz(&info.WorkflowExecution)
					c.Fuzz(&info.Type)
					c.Fuzz(&info.StartTime)
					c.Fuzz(&info.CloseTime)
					c.Fuzz(&info.CloseStatus) 
					c.Fuzz(&info.HistoryLength)
					c.Fuzz(&info.ExecutionTime)
					c.Fuzz(&info.TaskList)
					c.Fuzz(&info.IsCron)
					c.Fuzz(&info.CronOverlapPolicy) 
				},
			},
			ExcludedFields: []string{
				"TaskListInfo", // [BUG] TaskListInfo field is not mapping correctly between proto and thrift - becomes nil after round trip
				"UpdateTime", // [BUG] UpdateTime field is not mapping correctly between proto and thrift - becomes nil after round trip
				"PartitionConfig", // [BUG] PartitionConfig field is not mapping correctly between proto and thrift - becomes nil after round trip
			},
		},
	)
}
func TestWorkflowExecutionSignaledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionSignaledEventAttributes{nil, {}, &testdata.WorkflowExecutionSignaledEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionSignaledEventAttributes(thrift.WorkflowExecutionSignaledEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowExecutionSignaledEventAttributes,
		proto.WorkflowExecutionSignaledEventAttributes,
		FuzzOptions{
			ExcludedFields: []string{
				"RequestId", // [BUG] Field mapping issue - not being preserved correctly in mapper
			},
		},
	)
}
func TestWorkflowExecutionStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionStartedEventAttributes{nil, {}, &testdata.WorkflowExecutionStartedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionStartedEventAttributes(thrift.WorkflowExecutionStartedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowExecutionStartedEventAttributes,
		proto.WorkflowExecutionStartedEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(initiator *apiv1.ContinueAsNewInitiator, c fuzz.Continue) {
					validValues := []apiv1.ContinueAsNewInitiator{
						apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_INVALID,
						apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_DECIDER,
						apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_RETRY_POLICY,
						apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_CRON_SCHEDULE,
					}
					*initiator = validValues[c.Intn(len(validValues))]
				},
				func(kind *apiv1.TaskListKind, c fuzz.Continue) {
					validValues := []apiv1.TaskListKind{
						apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
						apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
						apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
					}
					*kind = validValues[c.Intn(len(validValues))]
				},
				func(attr *apiv1.WorkflowExecutionStartedEventAttributes, c fuzz.Continue) {
					c.Fuzz(&attr.WorkflowType)
					c.Fuzz(&attr.TaskList)
					c.Fuzz(&attr.Identity)
					c.Fuzz(&attr.FirstExecutionRunId)
					c.Fuzz(&attr.OriginalExecutionRunId)
				},
			},
			ExcludedFields: []string{
				// [BUG] ParentExecutionInfo has inconsistent behaviour - DomainId field is lost during round trip
				// The proto mapper panics with "either all or none parent execution info must be set"
				// TODO: Fix the ParentExecutionInfo mapping to preserve all fields consistently
				"ParentExecutionInfo",
			},
		},
	)
}
func TestWorkflowExecutionTerminatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionTerminatedEventAttributes{nil, {}, &testdata.WorkflowExecutionTerminatedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionTerminatedEventAttributes(thrift.WorkflowExecutionTerminatedEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowExecutionTerminatedEventAttributes,
		proto.WorkflowExecutionTerminatedEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkflowExecutionTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionTimedOutEventAttributes{nil, {}, &testdata.WorkflowExecutionTimedOutEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionTimedOutEventAttributes(thrift.WorkflowExecutionTimedOutEventAttributes(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowExecutionTimedOutEventAttributes,
		proto.WorkflowExecutionTimedOutEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.TimeoutType, c fuzz.Continue) {
					validValues := []apiv1.TimeoutType{
						apiv1.TimeoutType_TIMEOUT_TYPE_INVALID,
						apiv1.TimeoutType_TIMEOUT_TYPE_START_TO_CLOSE,
						apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_START,
						apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
						apiv1.TimeoutType_TIMEOUT_TYPE_HEARTBEAT,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestWorkflowQuery(t *testing.T) {
	for _, item := range []*apiv1.WorkflowQuery{nil, {}, &testdata.WorkflowQuery} {
		assert.Equal(t, item, proto.WorkflowQuery(thrift.WorkflowQuery(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowQuery,
		proto.WorkflowQuery,
		FuzzOptions{},
	)
}
func TestWorkflowQueryResult(t *testing.T) {
	for _, item := range []*apiv1.WorkflowQueryResult{nil, {}, &testdata.WorkflowQueryResult} {
		assert.Equal(t, item, proto.WorkflowQueryResult(thrift.WorkflowQueryResult(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowQueryResult,
		proto.WorkflowQueryResult,
		FuzzOptions{
			CustomFuncs: []interface{}{
				func(e *apiv1.QueryResultType, c fuzz.Continue) {
					validValues := []apiv1.QueryResultType{
						apiv1.QueryResultType_QUERY_RESULT_TYPE_INVALID,
						apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED,
						apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED,
					}
					*e = validValues[c.Intn(len(validValues))]
				},
			},
		},
	)
}
func TestWorkflowType(t *testing.T) {
	for _, item := range []*apiv1.WorkflowType{nil, {}, &testdata.WorkflowType} {
		assert.Equal(t, item, proto.WorkflowType(thrift.WorkflowType(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowType,
		proto.WorkflowType,
		FuzzOptions{},
	)
}
func TestWorkflowTypeFilter(t *testing.T) {
	for _, item := range []*apiv1.WorkflowTypeFilter{nil, {}, &testdata.WorkflowTypeFilter} {
		assert.Equal(t, item, proto.WorkflowTypeFilter(thrift.WorkflowTypeFilter(item)))
	}

	runFuzzTest(t,
		thrift.WorkflowTypeFilter,
		proto.WorkflowTypeFilter,
		FuzzOptions{},
	)
}
func TestDataBlobArray(t *testing.T) {
	for _, item := range [][]*apiv1.DataBlob{nil, {}, testdata.DataBlobArray} {
		assert.Equal(t, item, proto.DataBlobArray(thrift.DataBlobArray(item)))
	}
}
func TestHistoryEventArray(t *testing.T) {
	for _, item := range [][]*apiv1.HistoryEvent{nil, {}, testdata.HistoryEventArray} {
		assert.Equal(t, item, proto.HistoryEventArray(thrift.HistoryEventArray(item)))
	}
}
func TestTaskListPartitionMetadataArray(t *testing.T) {
	for _, item := range [][]*apiv1.TaskListPartitionMetadata{nil, {}, testdata.TaskListPartitionMetadataArray} {
		assert.Equal(t, item, proto.TaskListPartitionMetadataArray(thrift.TaskListPartitionMetadataArray(item)))
	}
}
func TestDecisionArray(t *testing.T) {
	for _, item := range [][]*apiv1.Decision{nil, {}, testdata.DecisionArray} {
		assert.Equal(t, item, proto.DecisionArray(thrift.DecisionArray(item)))
	}
}
func TestPollerInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.PollerInfo{nil, {}, testdata.PollerInfoArray} {
		assert.Equal(t, item, proto.PollerInfoArray(thrift.PollerInfoArray(item)))
	}
}
func TestPendingChildExecutionInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.PendingChildExecutionInfo{nil, {}, testdata.PendingChildExecutionInfoArray} {
		assert.Equal(t, item, proto.PendingChildExecutionInfoArray(thrift.PendingChildExecutionInfoArray(item)))
	}
}
func TestWorkflowExecutionInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.WorkflowExecutionInfo{nil, {}, testdata.WorkflowExecutionInfoArray} {
		assert.Equal(t, item, proto.WorkflowExecutionInfoArray(thrift.WorkflowExecutionInfoArray(item)))
	}
}
func TestDescribeDomainResponseArray(t *testing.T) {
	for _, item := range [][]*apiv1.Domain{nil, {}, testdata.DomainArray} {
		assert.Equal(t, item, proto.DescribeDomainResponseArray(thrift.DescribeDomainResponseArray(item)))
	}
}
func TestResetPointInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.ResetPointInfo{nil, {}, testdata.ResetPointInfoArray} {
		assert.Equal(t, item, proto.ResetPointInfoArray(thrift.ResetPointInfoArray(item)))
	}
}
func TestPendingActivityInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.PendingActivityInfo{nil, {}, testdata.PendingActivityInfoArray} {
		assert.Equal(t, item, proto.PendingActivityInfoArray(thrift.PendingActivityInfoArray(item)))
	}
}
func TestClusterReplicationConfigurationArray(t *testing.T) {
	for _, item := range [][]*apiv1.ClusterReplicationConfiguration{nil, {}, testdata.ClusterReplicationConfigurationArray} {
		assert.Equal(t, item, proto.ClusterReplicationConfigurationArray(thrift.ClusterReplicationConfigurationArray(item)))
	}
}
func TestActivityLocalDispatchInfoMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.ActivityLocalDispatchInfo{nil, {}, testdata.ActivityLocalDispatchInfoMap} {
		assert.Equal(t, item, proto.ActivityLocalDispatchInfoMap(thrift.ActivityLocalDispatchInfoMap(item)))
	}
}
func TestBadBinaryInfoMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.BadBinaryInfo{nil, {}, testdata.BadBinaryInfoMap} {
		assert.Equal(t, item, proto.BadBinaryInfoMap(thrift.BadBinaryInfoMap(item)))
	}
}
func TestIndexedValueTypeMap(t *testing.T) {
	for _, item := range []map[string]apiv1.IndexedValueType{nil, {}, testdata.IndexedValueTypeMap} {
		assert.Equal(t, item, proto.IndexedValueTypeMap(thrift.IndexedValueTypeMap(item)))
	}
}
func TestWorkflowQueryMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.WorkflowQuery{nil, {}, testdata.WorkflowQueryMap} {
		assert.Equal(t, item, proto.WorkflowQueryMap(thrift.WorkflowQueryMap(item)))
	}
}
func TestWorkflowQueryResultMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.WorkflowQueryResult{nil, {}, testdata.WorkflowQueryResultMap} {
		assert.Equal(t, item, proto.WorkflowQueryResultMap(thrift.WorkflowQueryResultMap(item)))
	}
}
func TestPayload(t *testing.T) {
	for _, item := range []*apiv1.Payload{nil, &testdata.Payload1} {
		assert.Equal(t, item, proto.Payload(thrift.Payload(item)))
	}

	assert.Equal(t, &apiv1.Payload{Data: []byte{}}, proto.Payload(thrift.Payload(&apiv1.Payload{})))

	runFuzzTest(t,
		thrift.Payload,
		proto.Payload,
		FuzzOptions{},
	)
}
func TestPayloadMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.Payload{nil, {}, testdata.PayloadMap} {
		assert.Equal(t, item, proto.PayloadMap(thrift.PayloadMap(item)))
	}
	for _, testObj := range testdata.PayloadMap {
		if testObj != nil {
		}
	}
}
func TestFailure(t *testing.T) {
	assert.Nil(t, proto.Failure(nil, nil))
	assert.Nil(t, thrift.FailureReason(nil))
	assert.Nil(t, thrift.FailureDetails(nil))
	failure := proto.Failure(&testdata.FailureReason, testdata.FailureDetails)
	assert.Equal(t, testdata.FailureReason, *thrift.FailureReason(failure))
	assert.Equal(t, testdata.FailureDetails, thrift.FailureDetails(failure))
}
func TestHistoryEvent(t *testing.T) {
	historyEvents := []*apiv1.HistoryEvent{
		nil,
		&testdata.HistoryEvent_WorkflowExecutionStarted,
		&testdata.HistoryEvent_WorkflowExecutionCompleted,
		&testdata.HistoryEvent_WorkflowExecutionFailed,
		&testdata.HistoryEvent_WorkflowExecutionTimedOut,
		&testdata.HistoryEvent_DecisionTaskScheduled,
		&testdata.HistoryEvent_DecisionTaskStarted,
		&testdata.HistoryEvent_DecisionTaskCompleted,
		&testdata.HistoryEvent_DecisionTaskTimedOut,
		&testdata.HistoryEvent_DecisionTaskFailed,
		&testdata.HistoryEvent_ActivityTaskScheduled,
		&testdata.HistoryEvent_ActivityTaskStarted,
		&testdata.HistoryEvent_ActivityTaskCompleted,
		&testdata.HistoryEvent_ActivityTaskFailed,
		&testdata.HistoryEvent_ActivityTaskTimedOut,
		&testdata.HistoryEvent_ActivityTaskCancelRequested,
		&testdata.HistoryEvent_RequestCancelActivityTaskFailed,
		&testdata.HistoryEvent_ActivityTaskCanceled,
		&testdata.HistoryEvent_TimerStarted,
		&testdata.HistoryEvent_TimerFired,
		&testdata.HistoryEvent_CancelTimerFailed,
		&testdata.HistoryEvent_TimerCanceled,
		&testdata.HistoryEvent_WorkflowExecutionCancelRequested,
		&testdata.HistoryEvent_WorkflowExecutionCanceled,
		&testdata.HistoryEvent_RequestCancelExternalWorkflowExecutionInitiated,
		&testdata.HistoryEvent_RequestCancelExternalWorkflowExecutionFailed,
		&testdata.HistoryEvent_ExternalWorkflowExecutionCancelRequested,
		&testdata.HistoryEvent_MarkerRecorded,
		&testdata.HistoryEvent_WorkflowExecutionSignaled,
		&testdata.HistoryEvent_WorkflowExecutionTerminated,
		&testdata.HistoryEvent_WorkflowExecutionContinuedAsNew,
		&testdata.HistoryEvent_StartChildWorkflowExecutionInitiated,
		&testdata.HistoryEvent_StartChildWorkflowExecutionFailed,
		&testdata.HistoryEvent_ChildWorkflowExecutionStarted,
		&testdata.HistoryEvent_ChildWorkflowExecutionCompleted,
		&testdata.HistoryEvent_ChildWorkflowExecutionFailed,
		&testdata.HistoryEvent_ChildWorkflowExecutionCanceled,
		&testdata.HistoryEvent_ChildWorkflowExecutionTimedOut,
		&testdata.HistoryEvent_ChildWorkflowExecutionTerminated,
		&testdata.HistoryEvent_SignalExternalWorkflowExecutionInitiated,
		&testdata.HistoryEvent_SignalExternalWorkflowExecutionFailed,
		&testdata.HistoryEvent_ExternalWorkflowExecutionSignaled,
		&testdata.HistoryEvent_UpsertWorkflowSearchAttributes,
	}
	for _, item := range historyEvents {
		assert.Equal(t, item, proto.HistoryEvent(thrift.HistoryEvent(item)))
	}
	assert.Panics(t, func() { proto.HistoryEvent(&shared.HistoryEvent{EventType: shared.EventType(UnknownValue).Ptr()}) })
	assert.Panics(t, func() { thrift.HistoryEvent(&apiv1.HistoryEvent{}) })
	// TODO: Add fuzz tests for TestHistoryEvent.
	// gofuzz struggles with the oneof types and the complex nature of a history event. With a sufficiently well written CustomFunc
	// it should be possible to add FuzzTesting to this file, though it'll require updating as/if the oneof changes.
}

func TestDecision(t *testing.T) {
	decisions := []*apiv1.Decision{
		nil,
		&testdata.Decision_CancelTimer,
		&testdata.Decision_CancelWorkflowExecution,
		&testdata.Decision_CompleteWorkflowExecution,
		&testdata.Decision_ContinueAsNewWorkflowExecution,
		&testdata.Decision_FailWorkflowExecution,
		&testdata.Decision_RecordMarker,
		&testdata.Decision_RequestCancelActivityTask,
		&testdata.Decision_RequestCancelExternalWorkflowExecution,
		&testdata.Decision_ScheduleActivityTask,
		&testdata.Decision_SignalExternalWorkflowExecution,
		&testdata.Decision_StartChildWorkflowExecution,
		&testdata.Decision_StartTimer,
		&testdata.Decision_UpsertWorkflowSearchAttributes,
	}
	for _, item := range decisions {
		assert.Equal(t, item, proto.Decision(thrift.Decision(item)))
	}
	assert.Panics(t, func() { proto.Decision(&shared.Decision{DecisionType: shared.DecisionType(UnknownValue).Ptr()}) })
	assert.Panics(t, func() { thrift.Decision(&apiv1.Decision{}) })
}
func TestListClosedWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListClosedWorkflowExecutionsRequest{
		nil,
		{},
		&testdata.ListClosedWorkflowExecutionsRequest_ExecutionFilter,
		&testdata.ListClosedWorkflowExecutionsRequest_StatusFilter,
		&testdata.ListClosedWorkflowExecutionsRequest_TypeFilter,
	} {
		assert.Equal(t, item, proto.ListClosedWorkflowExecutionsRequest(thrift.ListClosedWorkflowExecutionsRequest(item)))
	}
}
func TestListOpenWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListOpenWorkflowExecutionsRequest{
		nil,
		{},
		&testdata.ListOpenWorkflowExecutionsRequest_ExecutionFilter,
		&testdata.ListOpenWorkflowExecutionsRequest_TypeFilter,
	} {
		assert.Equal(t, item, proto.ListOpenWorkflowExecutionsRequest(thrift.ListOpenWorkflowExecutionsRequest(item)))
	}
}

// runFuzzTest provides a more type-safe version of runFuzzTest using generics
func runFuzzTest[TProto protobuf.Message, TThrift any](
	t *testing.T,
	protoToThrift func(TProto) TThrift,
	thriftToProto func(TThrift) TProto,
	options FuzzOptions,
) {
	// Apply defaults for zero values
	if options.NilChance == 0 {
		options.NilChance = DefaultNilChance
	}
	if options.Iterations == 0 {
		options.Iterations = DefaultIterations
	}

	// Build fuzzer functions - start with defaults and add custom ones
	fuzzerFuncs := []interface{}{
		// Default: Custom fuzzer for gogo protobuf timestamps
		func(ts *gogo.Timestamp, c fuzz.Continue) {
			ts.Seconds = c.Int63n(MaxSafeTimestampSeconds)
			ts.Nanos = c.Int31n(NanosecondsPerSecond)
		},
		// Default: Custom fuzzer for gogo protobuf durations
		// Note: Thrift protocol only supports second-level precision for durations,
		// so we only generate whole seconds to ensure round-trip compatibility
		func(d *gogo.Duration, c fuzz.Continue) {
			d.Seconds = c.Int63n(MaxDurationSeconds)
			d.Nanos = 0 // Thrift mapping truncates nanoseconds, so set to 0 for consistency
		},
		// Default: Custom fuzzer for Payload to handle data consistently
		func(p *apiv1.Payload, c fuzz.Continue) {
			length := c.Intn(MaxPayloadBytes) + 1 // 1-MaxPayloadBytes bytes
			p.Data = make([]byte, length)
			for i := 0; i < length; i++ {
				p.Data[i] = byte(c.Uint32())
			}
		},
	}

	// Add custom fuzzer functions
	fuzzerFuncs = append(fuzzerFuncs, options.CustomFuncs...)

	fuzzer := testdatagen.NewWithNilChance(t, int64(123), float32(options.NilChance), fuzzerFuncs...)

	for i := 0; i < options.Iterations; i++ {
		// Create new instance using generics
		var zero TProto
		fuzzed := reflect.New(reflect.TypeOf(zero).Elem()).Interface().(TProto)
		fuzzer.Fuzz(fuzzed)

		// Clear protobuf internal fields and apply field exclusions in a single pass
		clearProtobufAndExcludedFields(fuzzed, options.ExcludedFields)

		// Test proto -> thrift -> proto round trip
		thriftResult := protoToThrift(fuzzed)
		protoResult := thriftToProto(thriftResult)

		// Clear internal fields and excluded fields from result as well
		clearProtobufAndExcludedFields(protoResult, options.ExcludedFields)

		assert.Equal(t, fuzzed, protoResult, "Round trip failed for fuzzed data at iteration %d", i)
	}
}

// clearFieldsIf recursively traverses an object and clears fields that match the predicate function
func clearFieldsIf(obj interface{}, shouldClear func(fieldName string) bool) {
	if obj == nil {
		return
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := v.Type().Field(i).Name

		if shouldClear(fieldName) && field.CanSet() {
			field.Set(reflect.Zero(field.Type()))
		}

		// Recursively clear fields in nested structs and slices
		if field.CanInterface() {
			switch field.Kind() {
			case reflect.Ptr:
				if !field.IsNil() {
					clearFieldsIf(field.Interface(), shouldClear)
				}
			case reflect.Struct:
				clearFieldsIf(field.Addr().Interface(), shouldClear)
			case reflect.Slice:
				for j := 0; j < field.Len(); j++ {
					elem := field.Index(j)
					if elem.CanInterface() {
						clearFieldsIf(elem.Interface(), shouldClear)
					}
				}
			}
		}
	}
}

// clearProtobufAndExcludedFields combines protobuf clearing and field exclusion in a single pass
func clearProtobufAndExcludedFields(obj interface{}, excludedFields []string) {
	// Create a map for O(1) lookup of excluded fields
	excludedMap := make(map[string]bool)
	for _, field := range excludedFields {
		excludedMap[field] = true
	}

	clearFieldsIf(obj, func(fieldName string) bool {
		// Clear if it's a protobuf internal field OR if it's in the excluded list
		return strings.HasPrefix(fieldName, "XXX_") || excludedMap[fieldName]
	})
}
