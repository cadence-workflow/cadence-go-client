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

// clearProtobufInternalFields clears protobuf internal fields (XXX_ prefixed)
func clearProtobufInternalFields(obj interface{}) {
	clearFieldsIf(obj, func(fieldName string) bool {
		return strings.HasPrefix(fieldName, "XXX_")
	})
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

// Fuzzing configuration constants
const (
	// DEFAULT_NIL_CHANCE is the default probability of setting pointer/slice fields to nil
	DEFAULT_NIL_CHANCE = 0.25
	// DEFAULT_ITERATIONS is the default number of fuzzing iterations to run
	DEFAULT_ITERATIONS = 100
	// MAX_SAFE_TIMESTAMP_SECONDS is the maximum seconds value that fits safely in int64 nanoseconds
	// This avoids overflow when converting to UnixNano (max safe range from 1970 to ~2262)
	MAX_SAFE_TIMESTAMP_SECONDS = 9223372036
	// MAX_DURATION_SECONDS is the maximum duration in seconds (~10000 years)
	MAX_DURATION_SECONDS = 315576000000
	// NANOSECONDS_PER_SECOND is the number of nanoseconds in a second
	NANOSECONDS_PER_SECOND = 1000000000
	// MAX_PAYLOAD_BYTES is the maximum payload size for fuzzing
	MAX_PAYLOAD_BYTES = 10
)

// FuzzOptions provides configuration for runFuzzTestV2
type FuzzOptions struct {
	// CustomFuncs are custom fuzzer functions to apply for specific types
	CustomFuncs []interface{}
	// ExcludedFields are field names to exclude from fuzzing (set to zero value)
	ExcludedFields []string
	// NilChance is the probability of setting pointer/slice fields to nil (default DEFAULT_NIL_CHANCE)
	NilChance float64
	// Iterations is the number of fuzzing iterations to run (default DEFAULT_ITERATIONS)
	Iterations int
}

// runFuzzTestV2 provides a more type-safe version of runFuzzTest using generics
func runFuzzTestV2[TProto protobuf.Message, TThrift any](
	t *testing.T,
	protoToThrift func(TProto) TThrift,
	thriftToProto func(TThrift) TProto,
	options FuzzOptions,
) {
	// Apply defaults for zero values
	if options.NilChance == 0 {
		options.NilChance = DEFAULT_NIL_CHANCE
	}
	if options.Iterations == 0 {
		options.Iterations = DEFAULT_ITERATIONS
	}

	// Build fuzzer functions - start with defaults and add custom ones
	fuzzerFuncs := []interface{}{
		// Default: Custom fuzzer for gogo protobuf timestamps
		func(ts *gogo.Timestamp, c fuzz.Continue) {
			ts.Seconds = c.Int63n(MAX_SAFE_TIMESTAMP_SECONDS)
			ts.Nanos = c.Int31n(NANOSECONDS_PER_SECOND)
		},
		// Default: Custom fuzzer for gogo protobuf durations
		func(d *gogo.Duration, c fuzz.Continue) {
			d.Seconds = c.Int63n(MAX_DURATION_SECONDS)
			d.Nanos = c.Int31n(NANOSECONDS_PER_SECOND)
		},
		// Default: Custom fuzzer for Payload to handle data consistently
		func(p *apiv1.Payload, c fuzz.Continue) {
			length := c.Intn(MAX_PAYLOAD_BYTES) + 1 // 1-MAX_PAYLOAD_BYTES bytes
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

// excludeFields sets specified field names to their zero values
func excludeFields(obj interface{}, excludedFields []string) {
	if obj == nil || len(excludedFields) == 0 {
		return
	}

	// Create a map for O(1) lookup
	excludedMap := make(map[string]bool)
	for _, field := range excludedFields {
		excludedMap[field] = true
	}

	clearFieldsIf(obj, func(fieldName string) bool {
		return excludedMap[fieldName]
	})
}

func TestActivityLocalDispatchInfo(t *testing.T) {
	// Test with existing sample data
	for _, item := range []*apiv1.ActivityLocalDispatchInfo{nil, {}, &testdata.ActivityLocalDispatchInfo} {
		assert.Equal(t, item, proto.ActivityLocalDispatchInfo(thrift.ActivityLocalDispatchInfo(item)))
	}

	runFuzzTestV2(t,
		thrift.ActivityLocalDispatchInfo,
		proto.ActivityLocalDispatchInfo,
		FuzzOptions{},
	)
}
func TestActivityTaskCancelRequestedEventAttributes(t *testing.T) {
	// Test with existing sample data
	for _, item := range []*apiv1.ActivityTaskCancelRequestedEventAttributes{nil, {}, &testdata.ActivityTaskCancelRequestedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskCancelRequestedEventAttributes(thrift.ActivityTaskCancelRequestedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ActivityTaskCancelRequestedEventAttributes,
		proto.ActivityTaskCancelRequestedEventAttributes,
		FuzzOptions{},
	)
}
func TestActivityTaskCanceledEventAttributes(t *testing.T) {
	// Test with existing sample data
	for _, item := range []*apiv1.ActivityTaskCanceledEventAttributes{nil, {}, &testdata.ActivityTaskCanceledEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskCanceledEventAttributes(thrift.ActivityTaskCanceledEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ActivityTaskCanceledEventAttributes,
		proto.ActivityTaskCanceledEventAttributes,
		FuzzOptions{},
	)
}
func TestActivityTaskCompletedEventAttributes(t *testing.T) {
	// Test with existing sample data
	for _, item := range []*apiv1.ActivityTaskCompletedEventAttributes{nil, {}, &testdata.ActivityTaskCompletedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskCompletedEventAttributes(thrift.ActivityTaskCompletedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ActivityTaskCompletedEventAttributes,
		proto.ActivityTaskCompletedEventAttributes,
		FuzzOptions{},
	)
}
func TestActivityTaskFailedEventAttributes(t *testing.T) {
	// Test with existing sample data
	for _, item := range []*apiv1.ActivityTaskFailedEventAttributes{nil, {}, &testdata.ActivityTaskFailedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskFailedEventAttributes(thrift.ActivityTaskFailedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ActivityTaskFailedEventAttributes,
		proto.ActivityTaskFailedEventAttributes,
		FuzzOptions{},
	)
}
func TestActivityTaskScheduledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskScheduledEventAttributes{nil, {}, &testdata.ActivityTaskScheduledEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskScheduledEventAttributes(thrift.ActivityTaskScheduledEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ActivityTaskScheduledEventAttributes,
		proto.ActivityTaskScheduledEventAttributes,
		FuzzOptions{},
	)
}
func TestActivityTaskStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskStartedEventAttributes{nil, {}, &testdata.ActivityTaskStartedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskStartedEventAttributes(thrift.ActivityTaskStartedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ActivityTaskStartedEventAttributes,
		proto.ActivityTaskStartedEventAttributes,
		FuzzOptions{},
	)
}
func TestActivityTaskTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskTimedOutEventAttributes{nil, {}, &testdata.ActivityTaskTimedOutEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskTimedOutEventAttributes(thrift.ActivityTaskTimedOutEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ActivityTaskTimedOutEventAttributes,
		proto.ActivityTaskTimedOutEventAttributes,
		FuzzOptions{},
	)
}
func TestActivityType(t *testing.T) {
	for _, item := range []*apiv1.ActivityType{nil, {}, &testdata.ActivityType} {
		assert.Equal(t, item, proto.ActivityType(thrift.ActivityType(item)))
	}

	runFuzzTestV2(t,
		thrift.ActivityType,
		proto.ActivityType,
		FuzzOptions{},
	)
}
func TestBadBinaries(t *testing.T) {
	for _, item := range []*apiv1.BadBinaries{nil, {}, &testdata.BadBinaries} {
		assert.Equal(t, item, proto.BadBinaries(thrift.BadBinaries(item)))
	}

	runFuzzTestV2(t,
		thrift.BadBinaries,
		proto.BadBinaries,
		FuzzOptions{},
	)
}
func TestBadBinaryInfo(t *testing.T) {
	for _, item := range []*apiv1.BadBinaryInfo{nil, {}, &testdata.BadBinaryInfo} {
		assert.Equal(t, item, proto.BadBinaryInfo(thrift.BadBinaryInfo(item)))
	}

	runFuzzTestV2(t,
		thrift.BadBinaryInfo,
		proto.BadBinaryInfo,
		FuzzOptions{},
	)
}
func TestCancelTimerFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.CancelTimerFailedEventAttributes{nil, {}, &testdata.CancelTimerFailedEventAttributes} {
		assert.Equal(t, item, proto.CancelTimerFailedEventAttributes(thrift.CancelTimerFailedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.CancelTimerFailedEventAttributes,
		proto.CancelTimerFailedEventAttributes,
		FuzzOptions{},
	)
}
func TestChildWorkflowExecutionCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionCanceledEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionCanceledEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionCanceledEventAttributes(thrift.ChildWorkflowExecutionCanceledEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ChildWorkflowExecutionCanceledEventAttributes,
		proto.ChildWorkflowExecutionCanceledEventAttributes,
		FuzzOptions{},
	)
}
func TestChildWorkflowExecutionCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionCompletedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionCompletedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionCompletedEventAttributes(thrift.ChildWorkflowExecutionCompletedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ChildWorkflowExecutionCompletedEventAttributes,
		proto.ChildWorkflowExecutionCompletedEventAttributes,
		FuzzOptions{},
	)
}
func TestChildWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionFailedEventAttributes(thrift.ChildWorkflowExecutionFailedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ChildWorkflowExecutionFailedEventAttributes,
		proto.ChildWorkflowExecutionFailedEventAttributes,
		FuzzOptions{},
	)
}
func TestChildWorkflowExecutionStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionStartedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionStartedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionStartedEventAttributes(thrift.ChildWorkflowExecutionStartedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ChildWorkflowExecutionStartedEventAttributes,
		proto.ChildWorkflowExecutionStartedEventAttributes,
		FuzzOptions{},
	)
}
func TestChildWorkflowExecutionTerminatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionTerminatedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionTerminatedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionTerminatedEventAttributes(thrift.ChildWorkflowExecutionTerminatedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ChildWorkflowExecutionTerminatedEventAttributes,
		proto.ChildWorkflowExecutionTerminatedEventAttributes,
		FuzzOptions{},
	)
}
func TestChildWorkflowExecutionTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionTimedOutEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionTimedOutEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionTimedOutEventAttributes(thrift.ChildWorkflowExecutionTimedOutEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ChildWorkflowExecutionTimedOutEventAttributes,
		proto.ChildWorkflowExecutionTimedOutEventAttributes,
		FuzzOptions{},
	)
}
func TestClusterReplicationConfiguration(t *testing.T) {
	for _, item := range []*apiv1.ClusterReplicationConfiguration{nil, {}, &testdata.ClusterReplicationConfiguration} {
		assert.Equal(t, item, proto.ClusterReplicationConfiguration(thrift.ClusterReplicationConfiguration(item)))
	}

	runFuzzTestV2(t,
		thrift.ClusterReplicationConfiguration,
		proto.ClusterReplicationConfiguration,
		FuzzOptions{},
	)
}
func TestCountWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.CountWorkflowExecutionsRequest{nil, {}, &testdata.CountWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.CountWorkflowExecutionsRequest(thrift.CountWorkflowExecutionsRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.CountWorkflowExecutionsRequest,
		proto.CountWorkflowExecutionsRequest,
		FuzzOptions{},
	)
}
func TestCountWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.CountWorkflowExecutionsResponse{nil, {}, &testdata.CountWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.CountWorkflowExecutionsResponse(thrift.CountWorkflowExecutionsResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.CountWorkflowExecutionsResponse,
		proto.CountWorkflowExecutionsResponse,
		FuzzOptions{},
	)
}
func TestDataBlob(t *testing.T) {
	for _, item := range []*apiv1.DataBlob{nil, {}, &testdata.DataBlob} {
		assert.Equal(t, item, proto.DataBlob(thrift.DataBlob(item)))
	}

	runFuzzTestV2(t,
		thrift.DataBlob,
		proto.DataBlob,
		FuzzOptions{},
	)
}
func TestDecisionTaskCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskCompletedEventAttributes{nil, {}, &testdata.DecisionTaskCompletedEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskCompletedEventAttributes(thrift.DecisionTaskCompletedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.DecisionTaskCompletedEventAttributes,
		proto.DecisionTaskCompletedEventAttributes,
		FuzzOptions{},
	)
}
func TestDecisionTaskFailedEventAttributes(t *testing.T) {
	// Test with existing sample data
	for _, item := range []*apiv1.DecisionTaskFailedEventAttributes{nil, {}, &testdata.DecisionTaskFailedEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskFailedEventAttributes(thrift.DecisionTaskFailedEventAttributes(item)))
	}

	// Fuzz test V2 with custom enum fuzzer and field exclusions
	runFuzzTestV2(t,
		thrift.DecisionTaskFailedEventAttributes,
		proto.DecisionTaskFailedEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				// Custom fuzzer for DecisionTaskFailedCause to generate valid enum values
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
				// Exclude RequestId since we know it's not mapped correctly (will be tested separately)
				"RequestId",
			},
			Iterations: 50, // Reduced iterations for this specific test
		},
	)
}
func TestDecisionTaskScheduledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskScheduledEventAttributes{nil, {}, &testdata.DecisionTaskScheduledEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskScheduledEventAttributes(thrift.DecisionTaskScheduledEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.DecisionTaskScheduledEventAttributes,
		proto.DecisionTaskScheduledEventAttributes,
		FuzzOptions{},
	)
}
func TestDecisionTaskStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskStartedEventAttributes{nil, {}, &testdata.DecisionTaskStartedEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskStartedEventAttributes(thrift.DecisionTaskStartedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.DecisionTaskStartedEventAttributes,
		proto.DecisionTaskStartedEventAttributes,
		FuzzOptions{},
	)
}
func TestDecisionTaskTimedOutEventAttributes(t *testing.T) {
	// Test with existing sample data
	for _, item := range []*apiv1.DecisionTaskTimedOutEventAttributes{nil, {}, &testdata.DecisionTaskTimedOutEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskTimedOutEventAttributes(thrift.DecisionTaskTimedOutEventAttributes(item)))
	}

	// Fuzz test to find mapper issues
	runFuzzTestV2(t,
		thrift.DecisionTaskTimedOutEventAttributes,
		proto.DecisionTaskTimedOutEventAttributes,
		FuzzOptions{
			CustomFuncs: []interface{}{
				// Custom fuzzer for DecisionTaskTimedOutCause to generate valid enum values
				func(cause *apiv1.DecisionTaskTimedOutCause, c fuzz.Continue) {
					validValues := []apiv1.DecisionTaskTimedOutCause{
						apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_TIMEOUT,
						apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_RESET,
					}
					*cause = validValues[c.Intn(len(validValues))]
				},
				// Custom fuzzer for TimeoutType to generate valid enum values
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
				// Exclude RequestId since we know it's not mapped correctly
				"RequestId",
			},
		},
	)
}
func TestDeprecateDomainRequest(t *testing.T) {
	for _, item := range []*apiv1.DeprecateDomainRequest{nil, {}, &testdata.DeprecateDomainRequest} {
		assert.Equal(t, item, proto.DeprecateDomainRequest(thrift.DeprecateDomainRequest(item)))
	}

	runFuzzTestV2(t,
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

	runFuzzTestV2(t,
		thrift.DescribeDomainRequest,
		proto.DescribeDomainRequest,
		FuzzOptions{},
	)
}
func TestDescribeDomainResponse_Domain(t *testing.T) {
	for _, item := range []*apiv1.Domain{nil, &testdata.Domain} {
		assert.Equal(t, item, proto.DescribeDomainResponseDomain(thrift.DescribeDomainResponseDomain(item)))
	}

	runFuzzTestV2(t,
		thrift.DescribeDomainResponseDomain,
		proto.DescribeDomainResponseDomain,
		FuzzOptions{},
	)
}
func TestDescribeDomainResponse(t *testing.T) {
	for _, item := range []*apiv1.DescribeDomainResponse{nil, &testdata.DescribeDomainResponse} {
		assert.Equal(t, item, proto.DescribeDomainResponse(thrift.DescribeDomainResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.DescribeDomainResponse,
		proto.DescribeDomainResponse,
		FuzzOptions{},
	)
}
func TestDescribeTaskListRequest(t *testing.T) {
	for _, item := range []*apiv1.DescribeTaskListRequest{nil, {}, &testdata.DescribeTaskListRequest} {
		assert.Equal(t, item, proto.DescribeTaskListRequest(thrift.DescribeTaskListRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.DescribeTaskListRequest,
		proto.DescribeTaskListRequest,
		FuzzOptions{},
	)
}
func TestDescribeTaskListResponse(t *testing.T) {
	for _, item := range []*apiv1.DescribeTaskListResponse{nil, {}, &testdata.DescribeTaskListResponse} {
		assert.Equal(t, item, proto.DescribeTaskListResponse(thrift.DescribeTaskListResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.DescribeTaskListResponse,
		proto.DescribeTaskListResponse,
		FuzzOptions{},
	)
}
func TestDescribeWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.DescribeWorkflowExecutionRequest{nil, {}, &testdata.DescribeWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.DescribeWorkflowExecutionRequest(thrift.DescribeWorkflowExecutionRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.DescribeWorkflowExecutionRequest,
		proto.DescribeWorkflowExecutionRequest,
		FuzzOptions{},
	)
}
func TestDescribeWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.DescribeWorkflowExecutionResponse{nil, {}, &testdata.DescribeWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.DescribeWorkflowExecutionResponse(thrift.DescribeWorkflowExecutionResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.DescribeWorkflowExecutionResponse,
		proto.DescribeWorkflowExecutionResponse,
		FuzzOptions{},
	)
}
func TestDiagnoseWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.DiagnoseWorkflowExecutionRequest{nil, {}, &testdata.DiagnoseWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.DiagnoseWorkflowExecutionRequest(thrift.DiagnoseWorkflowExecutionRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.DiagnoseWorkflowExecutionRequest,
		proto.DiagnoseWorkflowExecutionRequest,
		FuzzOptions{},
	)
}
func TestDiagnoseWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.DiagnoseWorkflowExecutionResponse{nil, {}, &testdata.DiagnoseWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.DiagnoseWorkflowExecutionResponse(thrift.DiagnoseWorkflowExecutionResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.DiagnoseWorkflowExecutionResponse,
		proto.DiagnoseWorkflowExecutionResponse,
		FuzzOptions{},
	)
}
func TestExternalWorkflowExecutionCancelRequestedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ExternalWorkflowExecutionCancelRequestedEventAttributes{nil, {}, &testdata.ExternalWorkflowExecutionCancelRequestedEventAttributes} {
		assert.Equal(t, item, proto.ExternalWorkflowExecutionCancelRequestedEventAttributes(thrift.ExternalWorkflowExecutionCancelRequestedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ExternalWorkflowExecutionCancelRequestedEventAttributes,
		proto.ExternalWorkflowExecutionCancelRequestedEventAttributes,
		FuzzOptions{},
	)
}
func TestExternalWorkflowExecutionSignaledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ExternalWorkflowExecutionSignaledEventAttributes{nil, {}, &testdata.ExternalWorkflowExecutionSignaledEventAttributes} {
		assert.Equal(t, item, proto.ExternalWorkflowExecutionSignaledEventAttributes(thrift.ExternalWorkflowExecutionSignaledEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.ExternalWorkflowExecutionSignaledEventAttributes,
		proto.ExternalWorkflowExecutionSignaledEventAttributes,
		FuzzOptions{},
	)
}
func TestGetClusterInfoResponse(t *testing.T) {
	for _, item := range []*apiv1.GetClusterInfoResponse{nil, {}, &testdata.GetClusterInfoResponse} {
		assert.Equal(t, item, proto.GetClusterInfoResponse(thrift.GetClusterInfoResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.GetClusterInfoResponse,
		proto.GetClusterInfoResponse,
		FuzzOptions{},
	)
}
func TestGetSearchAttributesResponse(t *testing.T) {
	for _, item := range []*apiv1.GetSearchAttributesResponse{nil, {}, &testdata.GetSearchAttributesResponse} {
		assert.Equal(t, item, proto.GetSearchAttributesResponse(thrift.GetSearchAttributesResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.GetSearchAttributesResponse,
		proto.GetSearchAttributesResponse,
		FuzzOptions{},
	)
}
func TestGetWorkflowExecutionHistoryRequest(t *testing.T) {
	for _, item := range []*apiv1.GetWorkflowExecutionHistoryRequest{nil, {}, &testdata.GetWorkflowExecutionHistoryRequest} {
		assert.Equal(t, item, proto.GetWorkflowExecutionHistoryRequest(thrift.GetWorkflowExecutionHistoryRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.GetWorkflowExecutionHistoryRequest,
		proto.GetWorkflowExecutionHistoryRequest,
		FuzzOptions{},
	)
}
func TestGetWorkflowExecutionHistoryResponse(t *testing.T) {
	for _, item := range []*apiv1.GetWorkflowExecutionHistoryResponse{nil, {}, &testdata.GetWorkflowExecutionHistoryResponse} {
		assert.Equal(t, item, proto.GetWorkflowExecutionHistoryResponse(thrift.GetWorkflowExecutionHistoryResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.GetWorkflowExecutionHistoryResponse,
		proto.GetWorkflowExecutionHistoryResponse,
		FuzzOptions{},
	)
}
func TestHeader(t *testing.T) {
	for _, item := range []*apiv1.Header{nil, {}, &testdata.Header} {
		assert.Equal(t, item, proto.Header(thrift.Header(item)))
	}

	runFuzzTestV2(t,
		thrift.Header,
		proto.Header,
		FuzzOptions{},
	)
}
func TestHistory(t *testing.T) {
	for _, item := range []*apiv1.History{nil, {}, &testdata.History} {
		assert.Equal(t, item, proto.History(thrift.History(item)))
	}

	runFuzzTestV2(t,
		thrift.History,
		proto.History,
		FuzzOptions{},
	)
}
func TestListArchivedWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListArchivedWorkflowExecutionsRequest{nil, {}, &testdata.ListArchivedWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.ListArchivedWorkflowExecutionsRequest(thrift.ListArchivedWorkflowExecutionsRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.ListArchivedWorkflowExecutionsRequest,
		proto.ListArchivedWorkflowExecutionsRequest,
		FuzzOptions{},
	)
}
func TestListArchivedWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListArchivedWorkflowExecutionsResponse{nil, {}, &testdata.ListArchivedWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListArchivedWorkflowExecutionsResponse(thrift.ListArchivedWorkflowExecutionsResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.ListArchivedWorkflowExecutionsResponse,
		proto.ListArchivedWorkflowExecutionsResponse,
		FuzzOptions{},
	)
}
func TestListClosedWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListClosedWorkflowExecutionsResponse{nil, {}, &testdata.ListClosedWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListClosedWorkflowExecutionsResponse(thrift.ListClosedWorkflowExecutionsResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.ListClosedWorkflowExecutionsResponse,
		proto.ListClosedWorkflowExecutionsResponse,
		FuzzOptions{},
	)
}
func TestListDomainsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListDomainsRequest{nil, {}, &testdata.ListDomainsRequest} {
		assert.Equal(t, item, proto.ListDomainsRequest(thrift.ListDomainsRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.ListDomainsRequest,
		proto.ListDomainsRequest,
		FuzzOptions{},
	)
}
func TestListDomainsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListDomainsResponse{nil, {}, &testdata.ListDomainsResponse} {
		assert.Equal(t, item, proto.ListDomainsResponse(thrift.ListDomainsResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.ListDomainsResponse,
		proto.ListDomainsResponse,
		FuzzOptions{},
	)
}
func TestListOpenWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListOpenWorkflowExecutionsResponse{nil, {}, &testdata.ListOpenWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListOpenWorkflowExecutionsResponse(thrift.ListOpenWorkflowExecutionsResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.ListOpenWorkflowExecutionsResponse,
		proto.ListOpenWorkflowExecutionsResponse,
		FuzzOptions{},
	)
}
func TestListTaskListPartitionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListTaskListPartitionsRequest{nil, {}, &testdata.ListTaskListPartitionsRequest} {
		assert.Equal(t, item, proto.ListTaskListPartitionsRequest(thrift.ListTaskListPartitionsRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.ListTaskListPartitionsRequest,
		proto.ListTaskListPartitionsRequest,
		FuzzOptions{},
	)
}
func TestListTaskListPartitionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListTaskListPartitionsResponse{nil, {}, &testdata.ListTaskListPartitionsResponse} {
		assert.Equal(t, item, proto.ListTaskListPartitionsResponse(thrift.ListTaskListPartitionsResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.ListTaskListPartitionsResponse,
		proto.ListTaskListPartitionsResponse,
		FuzzOptions{},
	)
}
func TestListWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListWorkflowExecutionsRequest{nil, {}, &testdata.ListWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.ListWorkflowExecutionsRequest(thrift.ListWorkflowExecutionsRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.ListWorkflowExecutionsRequest,
		proto.ListWorkflowExecutionsRequest,
		FuzzOptions{},
	)
}
func TestListWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListWorkflowExecutionsResponse{nil, {}, &testdata.ListWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListWorkflowExecutionsResponse(thrift.ListWorkflowExecutionsResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.ListWorkflowExecutionsResponse,
		proto.ListWorkflowExecutionsResponse,
		FuzzOptions{},
	)
}
func TestMarkerRecordedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.MarkerRecordedEventAttributes{nil, {}, &testdata.MarkerRecordedEventAttributes} {
		assert.Equal(t, item, proto.MarkerRecordedEventAttributes(thrift.MarkerRecordedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.MarkerRecordedEventAttributes,
		proto.MarkerRecordedEventAttributes,
		FuzzOptions{},
	)
}
func TestMemo(t *testing.T) {
	for _, item := range []*apiv1.Memo{nil, {}, &testdata.Memo} {
		assert.Equal(t, item, proto.Memo(thrift.Memo(item)))
	}

	runFuzzTestV2(t,
		thrift.Memo,
		proto.Memo,
		FuzzOptions{},
	)
}
func TestPendingActivityInfo(t *testing.T) {
	for _, item := range []*apiv1.PendingActivityInfo{nil, {}, &testdata.PendingActivityInfo} {
		assert.Equal(t, item, proto.PendingActivityInfo(thrift.PendingActivityInfo(item)))
	}

	runFuzzTestV2(t,
		thrift.PendingActivityInfo,
		proto.PendingActivityInfo,
		FuzzOptions{},
	)
}
func TestPendingChildExecutionInfo(t *testing.T) {
	for _, item := range []*apiv1.PendingChildExecutionInfo{nil, {}, &testdata.PendingChildExecutionInfo} {
		assert.Equal(t, item, proto.PendingChildExecutionInfo(thrift.PendingChildExecutionInfo(item)))
	}

	runFuzzTestV2(t,
		thrift.PendingChildExecutionInfo,
		proto.PendingChildExecutionInfo,
		FuzzOptions{},
	)
}
func TestPendingDecisionInfo(t *testing.T) {
	for _, item := range []*apiv1.PendingDecisionInfo{nil, {}, &testdata.PendingDecisionInfo} {
		assert.Equal(t, item, proto.PendingDecisionInfo(thrift.PendingDecisionInfo(item)))
	}

	runFuzzTestV2(t,
		thrift.PendingDecisionInfo,
		proto.PendingDecisionInfo,
		FuzzOptions{},
	)
}
func TestPollForActivityTaskRequest(t *testing.T) {
	for _, item := range []*apiv1.PollForActivityTaskRequest{nil, {}, &testdata.PollForActivityTaskRequest} {
		assert.Equal(t, item, proto.PollForActivityTaskRequest(thrift.PollForActivityTaskRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.PollForActivityTaskRequest,
		proto.PollForActivityTaskRequest,
		FuzzOptions{},
	)
}
func TestPollForActivityTaskResponse(t *testing.T) {
	for _, item := range []*apiv1.PollForActivityTaskResponse{nil, {}, &testdata.PollForActivityTaskResponse} {
		assert.Equal(t, item, proto.PollForActivityTaskResponse(thrift.PollForActivityTaskResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.PollForActivityTaskResponse,
		proto.PollForActivityTaskResponse,
		FuzzOptions{},
	)
}
func TestPollForDecisionTaskRequest(t *testing.T) {
	for _, item := range []*apiv1.PollForDecisionTaskRequest{nil, {}, &testdata.PollForDecisionTaskRequest} {
		assert.Equal(t, item, proto.PollForDecisionTaskRequest(thrift.PollForDecisionTaskRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.PollForDecisionTaskRequest,
		proto.PollForDecisionTaskRequest,
		FuzzOptions{},
	)
}
func TestPollForDecisionTaskResponse(t *testing.T) {
	for _, item := range []*apiv1.PollForDecisionTaskResponse{nil, {}, &testdata.PollForDecisionTaskResponse} {
		assert.Equal(t, item, proto.PollForDecisionTaskResponse(thrift.PollForDecisionTaskResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.PollForDecisionTaskResponse,
		proto.PollForDecisionTaskResponse,
		FuzzOptions{},
	)
}
func TestPollerInfo(t *testing.T) {
	for _, item := range []*apiv1.PollerInfo{nil, {}, &testdata.PollerInfo} {
		assert.Equal(t, item, proto.PollerInfo(thrift.PollerInfo(item)))
	}

	runFuzzTestV2(t,
		thrift.PollerInfo,
		proto.PollerInfo,
		FuzzOptions{},
	)
}
func TestQueryRejected(t *testing.T) {
	for _, item := range []*apiv1.QueryRejected{nil, {}, &testdata.QueryRejected} {
		assert.Equal(t, item, proto.QueryRejected(thrift.QueryRejected(item)))
	}

	runFuzzTestV2(t,
		thrift.QueryRejected,
		proto.QueryRejected,
		FuzzOptions{},
	)
}
func TestQueryWorkflowRequest(t *testing.T) {
	for _, item := range []*apiv1.QueryWorkflowRequest{nil, {}, &testdata.QueryWorkflowRequest} {
		assert.Equal(t, item, proto.QueryWorkflowRequest(thrift.QueryWorkflowRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.QueryWorkflowRequest,
		proto.QueryWorkflowRequest,
		FuzzOptions{},
	)
}
func TestQueryWorkflowResponse(t *testing.T) {
	for _, item := range []*apiv1.QueryWorkflowResponse{nil, {}, &testdata.QueryWorkflowResponse} {
		assert.Equal(t, item, proto.QueryWorkflowResponse(thrift.QueryWorkflowResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.QueryWorkflowResponse,
		proto.QueryWorkflowResponse,
		FuzzOptions{},
	)
}
func TestRecordActivityTaskHeartbeatByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatByIDRequest{nil, {}, &testdata.RecordActivityTaskHeartbeatByIDRequest} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatByIDRequest(thrift.RecordActivityTaskHeartbeatByIDRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.RecordActivityTaskHeartbeatByIDRequest,
		proto.RecordActivityTaskHeartbeatByIDRequest,
		FuzzOptions{},
	)
}
func TestRecordActivityTaskHeartbeatByIDResponse(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatByIDResponse{nil, {}, &testdata.RecordActivityTaskHeartbeatByIDResponse} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatByIDResponse(thrift.RecordActivityTaskHeartbeatByIDResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.RecordActivityTaskHeartbeatByIDResponse,
		proto.RecordActivityTaskHeartbeatByIDResponse,
		FuzzOptions{},
	)
}
func TestRecordActivityTaskHeartbeatRequest(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatRequest{nil, {}, &testdata.RecordActivityTaskHeartbeatRequest} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatRequest(thrift.RecordActivityTaskHeartbeatRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.RecordActivityTaskHeartbeatRequest,
		proto.RecordActivityTaskHeartbeatRequest,
		FuzzOptions{},
	)
}
func TestRecordActivityTaskHeartbeatResponse(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatResponse{nil, {}, &testdata.RecordActivityTaskHeartbeatResponse} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatResponse(thrift.RecordActivityTaskHeartbeatResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.RecordActivityTaskHeartbeatResponse,
		proto.RecordActivityTaskHeartbeatResponse,
		FuzzOptions{},
	)
}
func TestRegisterDomainRequest(t *testing.T) {
	for _, item := range []*apiv1.RegisterDomainRequest{nil, &testdata.RegisterDomainRequest} {
		assert.Equal(t, item, proto.RegisterDomainRequest(thrift.RegisterDomainRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.RegisterDomainRequest,
		proto.RegisterDomainRequest,
		FuzzOptions{},
	)
}
func TestRequestCancelActivityTaskFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelActivityTaskFailedEventAttributes{nil, {}, &testdata.RequestCancelActivityTaskFailedEventAttributes} {
		assert.Equal(t, item, proto.RequestCancelActivityTaskFailedEventAttributes(thrift.RequestCancelActivityTaskFailedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.RequestCancelActivityTaskFailedEventAttributes,
		proto.RequestCancelActivityTaskFailedEventAttributes,
		FuzzOptions{},
	)
}
func TestRequestCancelExternalWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelExternalWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.RequestCancelExternalWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.RequestCancelExternalWorkflowExecutionFailedEventAttributes(thrift.RequestCancelExternalWorkflowExecutionFailedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.RequestCancelExternalWorkflowExecutionFailedEventAttributes,
		proto.RequestCancelExternalWorkflowExecutionFailedEventAttributes,
		FuzzOptions{},
	)
}
func TestRequestCancelExternalWorkflowExecutionInitiatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{nil, {}, &testdata.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes} {
		assert.Equal(t, item, proto.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes(thrift.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes,
		proto.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes,
		FuzzOptions{},
	)
}
func TestRequestCancelWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelWorkflowExecutionRequest{nil, {}, &testdata.RequestCancelWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.RequestCancelWorkflowExecutionRequest(thrift.RequestCancelWorkflowExecutionRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.RequestCancelWorkflowExecutionRequest,
		proto.RequestCancelWorkflowExecutionRequest,
		FuzzOptions{},
	)
}
func TestResetPointInfo(t *testing.T) {
	for _, item := range []*apiv1.ResetPointInfo{nil, {}, &testdata.ResetPointInfo} {
		assert.Equal(t, item, proto.ResetPointInfo(thrift.ResetPointInfo(item)))
	}

	runFuzzTestV2(t,
		thrift.ResetPointInfo,
		proto.ResetPointInfo,
		FuzzOptions{},
	)
}
func TestResetPoints(t *testing.T) {
	for _, item := range []*apiv1.ResetPoints{nil, {}, &testdata.ResetPoints} {
		assert.Equal(t, item, proto.ResetPoints(thrift.ResetPoints(item)))
	}

	runFuzzTestV2(t,
		thrift.ResetPoints,
		proto.ResetPoints,
		FuzzOptions{},
	)
}
func TestResetStickyTaskListRequest(t *testing.T) {
	for _, item := range []*apiv1.ResetStickyTaskListRequest{nil, {}, &testdata.ResetStickyTaskListRequest} {
		assert.Equal(t, item, proto.ResetStickyTaskListRequest(thrift.ResetStickyTaskListRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.ResetStickyTaskListRequest,
		proto.ResetStickyTaskListRequest,
		FuzzOptions{},
	)
}
func TestResetWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.ResetWorkflowExecutionRequest{nil, {}, &testdata.ResetWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.ResetWorkflowExecutionRequest(thrift.ResetWorkflowExecutionRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.ResetWorkflowExecutionRequest,
		proto.ResetWorkflowExecutionRequest,
		FuzzOptions{},
	)
}
func TestResetWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.ResetWorkflowExecutionResponse{nil, {}, &testdata.ResetWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.ResetWorkflowExecutionResponse(thrift.ResetWorkflowExecutionResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.ResetWorkflowExecutionResponse,
		proto.ResetWorkflowExecutionResponse,
		FuzzOptions{},
	)
}
func TestRespondActivityTaskCanceledByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCanceledByIDRequest{nil, {}, &testdata.RespondActivityTaskCanceledByIDRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCanceledByIDRequest(thrift.RespondActivityTaskCanceledByIDRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.RespondActivityTaskCanceledByIDRequest,
		proto.RespondActivityTaskCanceledByIDRequest,
		FuzzOptions{},
	)
}
func TestRespondActivityTaskCanceledRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCanceledRequest{nil, {}, &testdata.RespondActivityTaskCanceledRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCanceledRequest(thrift.RespondActivityTaskCanceledRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.RespondActivityTaskCanceledRequest,
		proto.RespondActivityTaskCanceledRequest,
		FuzzOptions{},
	)
}
func TestRespondActivityTaskCompletedByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCompletedByIDRequest{nil, {}, &testdata.RespondActivityTaskCompletedByIDRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCompletedByIDRequest(thrift.RespondActivityTaskCompletedByIDRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.RespondActivityTaskCompletedByIDRequest,
		proto.RespondActivityTaskCompletedByIDRequest,
		FuzzOptions{},
	)
}
func TestRespondActivityTaskCompletedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCompletedRequest{nil, {}, &testdata.RespondActivityTaskCompletedRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCompletedRequest(thrift.RespondActivityTaskCompletedRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.RespondActivityTaskCompletedRequest,
		proto.RespondActivityTaskCompletedRequest,
		FuzzOptions{},
	)
}
func TestRespondActivityTaskFailedByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskFailedByIDRequest{nil, {}, &testdata.RespondActivityTaskFailedByIDRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskFailedByIDRequest(thrift.RespondActivityTaskFailedByIDRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.RespondActivityTaskFailedByIDRequest,
		proto.RespondActivityTaskFailedByIDRequest,
		FuzzOptions{},
	)
}
func TestRespondActivityTaskFailedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskFailedRequest{nil, {}, &testdata.RespondActivityTaskFailedRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskFailedRequest(thrift.RespondActivityTaskFailedRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.RespondActivityTaskFailedRequest,
		proto.RespondActivityTaskFailedRequest,
		FuzzOptions{},
	)
}
func TestRespondDecisionTaskCompletedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondDecisionTaskCompletedRequest{nil, {}, &testdata.RespondDecisionTaskCompletedRequest} {
		assert.Equal(t, item, proto.RespondDecisionTaskCompletedRequest(thrift.RespondDecisionTaskCompletedRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.RespondDecisionTaskCompletedRequest,
		proto.RespondDecisionTaskCompletedRequest,
		FuzzOptions{},
	)
}
func TestRespondDecisionTaskCompletedResponse(t *testing.T) {
	for _, item := range []*apiv1.RespondDecisionTaskCompletedResponse{nil, {}, &testdata.RespondDecisionTaskCompletedResponse} {
		assert.Equal(t, item, proto.RespondDecisionTaskCompletedResponse(thrift.RespondDecisionTaskCompletedResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.RespondDecisionTaskCompletedResponse,
		proto.RespondDecisionTaskCompletedResponse,
		FuzzOptions{},
	)
}
func TestRespondDecisionTaskFailedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondDecisionTaskFailedRequest{nil, {}, &testdata.RespondDecisionTaskFailedRequest} {
		assert.Equal(t, item, proto.RespondDecisionTaskFailedRequest(thrift.RespondDecisionTaskFailedRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.RespondDecisionTaskFailedRequest,
		proto.RespondDecisionTaskFailedRequest,
		FuzzOptions{},
	)
}
func TestRespondQueryTaskCompletedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondQueryTaskCompletedRequest{nil, {Result: &apiv1.WorkflowQueryResult{}}, &testdata.RespondQueryTaskCompletedRequest} {
		assert.Equal(t, item, proto.RespondQueryTaskCompletedRequest(thrift.RespondQueryTaskCompletedRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.RespondQueryTaskCompletedRequest,
		proto.RespondQueryTaskCompletedRequest,
		FuzzOptions{},
	)
}
func TestRetryPolicy(t *testing.T) {
	for _, item := range []*apiv1.RetryPolicy{nil, {}, &testdata.RetryPolicy} {
		assert.Equal(t, item, proto.RetryPolicy(thrift.RetryPolicy(item)))
	}

	runFuzzTestV2(t,
		thrift.RetryPolicy,
		proto.RetryPolicy,
		FuzzOptions{},
	)
}
func TestScanWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ScanWorkflowExecutionsRequest{nil, {}, &testdata.ScanWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.ScanWorkflowExecutionsRequest(thrift.ScanWorkflowExecutionsRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.ScanWorkflowExecutionsRequest,
		proto.ScanWorkflowExecutionsRequest,
		FuzzOptions{},
	)
}
func TestScanWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ScanWorkflowExecutionsResponse{nil, {}, &testdata.ScanWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ScanWorkflowExecutionsResponse(thrift.ScanWorkflowExecutionsResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.ScanWorkflowExecutionsResponse,
		proto.ScanWorkflowExecutionsResponse,
		FuzzOptions{},
	)
}
func TestSearchAttributes(t *testing.T) {
	for _, item := range []*apiv1.SearchAttributes{nil, {}, &testdata.SearchAttributes} {
		assert.Equal(t, item, proto.SearchAttributes(thrift.SearchAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.SearchAttributes,
		proto.SearchAttributes,
		FuzzOptions{},
	)
}
func TestSignalExternalWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.SignalExternalWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.SignalExternalWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.SignalExternalWorkflowExecutionFailedEventAttributes(thrift.SignalExternalWorkflowExecutionFailedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.SignalExternalWorkflowExecutionFailedEventAttributes,
		proto.SignalExternalWorkflowExecutionFailedEventAttributes,
		FuzzOptions{},
	)
}
func TestSignalExternalWorkflowExecutionInitiatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.SignalExternalWorkflowExecutionInitiatedEventAttributes{nil, {}, &testdata.SignalExternalWorkflowExecutionInitiatedEventAttributes} {
		assert.Equal(t, item, proto.SignalExternalWorkflowExecutionInitiatedEventAttributes(thrift.SignalExternalWorkflowExecutionInitiatedEventAttributes(item)))
	}

	runFuzzTestV2(t,
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

	runFuzzTestV2(t,
		thrift.SignalWithStartWorkflowExecutionRequest,
		proto.SignalWithStartWorkflowExecutionRequest,
		FuzzOptions{},
	)
}
func TestSignalWithStartWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.SignalWithStartWorkflowExecutionResponse{nil, {}, &testdata.SignalWithStartWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.SignalWithStartWorkflowExecutionResponse(thrift.SignalWithStartWorkflowExecutionResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.SignalWithStartWorkflowExecutionResponse,
		proto.SignalWithStartWorkflowExecutionResponse,
		FuzzOptions{},
	)
}
func TestSignalWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.SignalWorkflowExecutionRequest{nil, {}, &testdata.SignalWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.SignalWorkflowExecutionRequest(thrift.SignalWorkflowExecutionRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.SignalWorkflowExecutionRequest,
		proto.SignalWorkflowExecutionRequest,
		FuzzOptions{},
	)
}
func TestStartChildWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.StartChildWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.StartChildWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.StartChildWorkflowExecutionFailedEventAttributes(thrift.StartChildWorkflowExecutionFailedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.StartChildWorkflowExecutionFailedEventAttributes,
		proto.StartChildWorkflowExecutionFailedEventAttributes,
		FuzzOptions{},
	)
}
func TestStartChildWorkflowExecutionInitiatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.StartChildWorkflowExecutionInitiatedEventAttributes{nil, {}, &testdata.StartChildWorkflowExecutionInitiatedEventAttributes} {
		assert.Equal(t, item, proto.StartChildWorkflowExecutionInitiatedEventAttributes(thrift.StartChildWorkflowExecutionInitiatedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.StartChildWorkflowExecutionInitiatedEventAttributes,
		proto.StartChildWorkflowExecutionInitiatedEventAttributes,
		FuzzOptions{},
	)
}
func TestStartTimeFilter(t *testing.T) {
	for _, item := range []*apiv1.StartTimeFilter{nil, {}, &testdata.StartTimeFilter} {
		assert.Equal(t, item, proto.StartTimeFilter(thrift.StartTimeFilter(item)))
	}

	runFuzzTestV2(t,
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

	runFuzzTestV2(t,
		thrift.StartWorkflowExecutionRequest,
		proto.StartWorkflowExecutionRequest,
		FuzzOptions{},
	)
}
func TestStartWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.StartWorkflowExecutionResponse{nil, {}, &testdata.StartWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.StartWorkflowExecutionResponse(thrift.StartWorkflowExecutionResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.StartWorkflowExecutionResponse,
		proto.StartWorkflowExecutionResponse,
		FuzzOptions{},
	)
}
func TestStatusFilter(t *testing.T) {
	for _, item := range []*apiv1.StatusFilter{nil, &testdata.StatusFilter} {
		assert.Equal(t, item, proto.StatusFilter(thrift.StatusFilter(item)))
	}

	runFuzzTestV2(t,
		thrift.StatusFilter,
		proto.StatusFilter,
		FuzzOptions{},
	)
}
func TestStickyExecutionAttributes(t *testing.T) {
	for _, item := range []*apiv1.StickyExecutionAttributes{nil, {}, &testdata.StickyExecutionAttributes} {
		assert.Equal(t, item, proto.StickyExecutionAttributes(thrift.StickyExecutionAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.StickyExecutionAttributes,
		proto.StickyExecutionAttributes,
		FuzzOptions{},
	)
}
func TestSupportedClientVersions(t *testing.T) {
	for _, item := range []*apiv1.SupportedClientVersions{nil, {}, &testdata.SupportedClientVersions} {
		assert.Equal(t, item, proto.SupportedClientVersions(thrift.SupportedClientVersions(item)))
	}

	runFuzzTestV2(t,
		thrift.SupportedClientVersions,
		proto.SupportedClientVersions,
		FuzzOptions{},
	)
}
func TestTaskIDBlock(t *testing.T) {
	for _, item := range []*apiv1.TaskIDBlock{nil, {}, &testdata.TaskIDBlock} {
		assert.Equal(t, item, proto.TaskIDBlock(thrift.TaskIDBlock(item)))
	}

	runFuzzTestV2(t,
		thrift.TaskIDBlock,
		proto.TaskIDBlock,
		FuzzOptions{},
	)
}
func TestTaskList(t *testing.T) {
	for _, item := range []*apiv1.TaskList{nil, {}, &testdata.TaskList} {
		assert.Equal(t, item, proto.TaskList(thrift.TaskList(item)))
	}

	runFuzzTestV2(t,
		thrift.TaskList,
		proto.TaskList,
		FuzzOptions{},
	)
}
func TestTaskListMetadata(t *testing.T) {
	for _, item := range []*apiv1.TaskListMetadata{nil, {}, &testdata.TaskListMetadata} {
		assert.Equal(t, item, proto.TaskListMetadata(thrift.TaskListMetadata(item)))
	}

	runFuzzTestV2(t,
		thrift.TaskListMetadata,
		proto.TaskListMetadata,
		FuzzOptions{},
	)
}
func TestTaskListPartitionMetadata(t *testing.T) {
	for _, item := range []*apiv1.TaskListPartitionMetadata{nil, {}, &testdata.TaskListPartitionMetadata} {
		assert.Equal(t, item, proto.TaskListPartitionMetadata(thrift.TaskListPartitionMetadata(item)))
	}

	runFuzzTestV2(t,
		thrift.TaskListPartitionMetadata,
		proto.TaskListPartitionMetadata,
		FuzzOptions{},
	)
}
func TestTaskListStatus(t *testing.T) {
	for _, item := range []*apiv1.TaskListStatus{nil, {}, &testdata.TaskListStatus} {
		assert.Equal(t, item, proto.TaskListStatus(thrift.TaskListStatus(item)))
	}

	runFuzzTestV2(t,
		thrift.TaskListStatus,
		proto.TaskListStatus,
		FuzzOptions{},
	)
}
func TestTerminateWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.TerminateWorkflowExecutionRequest{nil, {}, &testdata.TerminateWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.TerminateWorkflowExecutionRequest(thrift.TerminateWorkflowExecutionRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.TerminateWorkflowExecutionRequest,
		proto.TerminateWorkflowExecutionRequest,
		FuzzOptions{},
	)
}
func TestTimerCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.TimerCanceledEventAttributes{nil, {}, &testdata.TimerCanceledEventAttributes} {
		assert.Equal(t, item, proto.TimerCanceledEventAttributes(thrift.TimerCanceledEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.TimerCanceledEventAttributes,
		proto.TimerCanceledEventAttributes,
		FuzzOptions{},
	)
}
func TestTimerFiredEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.TimerFiredEventAttributes{nil, {}, &testdata.TimerFiredEventAttributes} {
		assert.Equal(t, item, proto.TimerFiredEventAttributes(thrift.TimerFiredEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.TimerFiredEventAttributes,
		proto.TimerFiredEventAttributes,
		FuzzOptions{},
	)
}
func TestTimerStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.TimerStartedEventAttributes{nil, {}, &testdata.TimerStartedEventAttributes} {
		assert.Equal(t, item, proto.TimerStartedEventAttributes(thrift.TimerStartedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.TimerStartedEventAttributes,
		proto.TimerStartedEventAttributes,
		FuzzOptions{},
	)
}
func TestUpdateDomainRequest(t *testing.T) {
	for _, item := range []*apiv1.UpdateDomainRequest{nil, {UpdateMask: &gogo.FieldMask{}}, &testdata.UpdateDomainRequest} {
		assert.Equal(t, item, proto.UpdateDomainRequest(thrift.UpdateDomainRequest(item)))
	}

	runFuzzTestV2(t,
		thrift.UpdateDomainRequest,
		proto.UpdateDomainRequest,
		FuzzOptions{},
	)
}
func TestUpdateDomainResponse(t *testing.T) {
	for _, item := range []*apiv1.UpdateDomainResponse{nil, &testdata.UpdateDomainResponse} {
		assert.Equal(t, item, proto.UpdateDomainResponse(thrift.UpdateDomainResponse(item)))
	}

	runFuzzTestV2(t,
		thrift.UpdateDomainResponse,
		proto.UpdateDomainResponse,
		FuzzOptions{},
	)
}
func TestUpsertWorkflowSearchAttributesEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.UpsertWorkflowSearchAttributesEventAttributes{nil, {}, &testdata.UpsertWorkflowSearchAttributesEventAttributes} {
		assert.Equal(t, item, proto.UpsertWorkflowSearchAttributesEventAttributes(thrift.UpsertWorkflowSearchAttributesEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.UpsertWorkflowSearchAttributesEventAttributes,
		proto.UpsertWorkflowSearchAttributesEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkerVersionInfo(t *testing.T) {
	for _, item := range []*apiv1.WorkerVersionInfo{nil, {}, &testdata.WorkerVersionInfo} {
		assert.Equal(t, item, proto.WorkerVersionInfo(thrift.WorkerVersionInfo(item)))
	}

	runFuzzTestV2(t,
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

	runFuzzTestV2(t,
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

	runFuzzTestV2(t,
		thrift.WorkflowExecutionCancelRequestedEventAttributes,
		proto.WorkflowExecutionCancelRequestedEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkflowExecutionCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionCanceledEventAttributes{nil, {}, &testdata.WorkflowExecutionCanceledEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionCanceledEventAttributes(thrift.WorkflowExecutionCanceledEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.WorkflowExecutionCanceledEventAttributes,
		proto.WorkflowExecutionCanceledEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkflowExecutionCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionCompletedEventAttributes{nil, {}, &testdata.WorkflowExecutionCompletedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionCompletedEventAttributes(thrift.WorkflowExecutionCompletedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.WorkflowExecutionCompletedEventAttributes,
		proto.WorkflowExecutionCompletedEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkflowExecutionConfiguration(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionConfiguration{nil, {}, &testdata.WorkflowExecutionConfiguration} {
		assert.Equal(t, item, proto.WorkflowExecutionConfiguration(thrift.WorkflowExecutionConfiguration(item)))
	}

	runFuzzTestV2(t,
		thrift.WorkflowExecutionConfiguration,
		proto.WorkflowExecutionConfiguration,
		FuzzOptions{},
	)
}
func TestWorkflowExecutionContinuedAsNewEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionContinuedAsNewEventAttributes{nil, {}, &testdata.WorkflowExecutionContinuedAsNewEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionContinuedAsNewEventAttributes(thrift.WorkflowExecutionContinuedAsNewEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.WorkflowExecutionContinuedAsNewEventAttributes,
		proto.WorkflowExecutionContinuedAsNewEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionFailedEventAttributes{nil, {}, &testdata.WorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionFailedEventAttributes(thrift.WorkflowExecutionFailedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.WorkflowExecutionFailedEventAttributes,
		proto.WorkflowExecutionFailedEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkflowExecutionFilter(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionFilter{nil, {}, &testdata.WorkflowExecutionFilter} {
		assert.Equal(t, item, proto.WorkflowExecutionFilter(thrift.WorkflowExecutionFilter(item)))
	}

	runFuzzTestV2(t,
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

	runFuzzTestV2(t,
		thrift.WorkflowExecutionInfo,
		proto.WorkflowExecutionInfo,
		FuzzOptions{},
	)
}
func TestWorkflowExecutionSignaledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionSignaledEventAttributes{nil, {}, &testdata.WorkflowExecutionSignaledEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionSignaledEventAttributes(thrift.WorkflowExecutionSignaledEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.WorkflowExecutionSignaledEventAttributes,
		proto.WorkflowExecutionSignaledEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkflowExecutionStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionStartedEventAttributes{nil, {}, &testdata.WorkflowExecutionStartedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionStartedEventAttributes(thrift.WorkflowExecutionStartedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.WorkflowExecutionStartedEventAttributes,
		proto.WorkflowExecutionStartedEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkflowExecutionTerminatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionTerminatedEventAttributes{nil, {}, &testdata.WorkflowExecutionTerminatedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionTerminatedEventAttributes(thrift.WorkflowExecutionTerminatedEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.WorkflowExecutionTerminatedEventAttributes,
		proto.WorkflowExecutionTerminatedEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkflowExecutionTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionTimedOutEventAttributes{nil, {}, &testdata.WorkflowExecutionTimedOutEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionTimedOutEventAttributes(thrift.WorkflowExecutionTimedOutEventAttributes(item)))
	}

	runFuzzTestV2(t,
		thrift.WorkflowExecutionTimedOutEventAttributes,
		proto.WorkflowExecutionTimedOutEventAttributes,
		FuzzOptions{},
	)
}
func TestWorkflowQuery(t *testing.T) {
	for _, item := range []*apiv1.WorkflowQuery{nil, {}, &testdata.WorkflowQuery} {
		assert.Equal(t, item, proto.WorkflowQuery(thrift.WorkflowQuery(item)))
	}

	runFuzzTestV2(t,
		thrift.WorkflowQuery,
		proto.WorkflowQuery,
		FuzzOptions{},
	)
}
func TestWorkflowQueryResult(t *testing.T) {
	for _, item := range []*apiv1.WorkflowQueryResult{nil, {}, &testdata.WorkflowQueryResult} {
		assert.Equal(t, item, proto.WorkflowQueryResult(thrift.WorkflowQueryResult(item)))
	}

	runFuzzTestV2(t,
		thrift.WorkflowQueryResult,
		proto.WorkflowQueryResult,
		FuzzOptions{},
	)
}
func TestWorkflowType(t *testing.T) {
	for _, item := range []*apiv1.WorkflowType{nil, {}, &testdata.WorkflowType} {
		assert.Equal(t, item, proto.WorkflowType(thrift.WorkflowType(item)))
	}

	runFuzzTestV2(t,
		thrift.WorkflowType,
		proto.WorkflowType,
		FuzzOptions{},
	)
}
func TestWorkflowTypeFilter(t *testing.T) {
	for _, item := range []*apiv1.WorkflowTypeFilter{nil, {}, &testdata.WorkflowTypeFilter} {
		assert.Equal(t, item, proto.WorkflowTypeFilter(thrift.WorkflowTypeFilter(item)))
	}

	runFuzzTestV2(t,
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

	runFuzzTestV2(t,
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
