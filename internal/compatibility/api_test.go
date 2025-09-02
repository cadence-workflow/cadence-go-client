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

	gogo "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/compatibility/proto"
	"go.uber.org/cadence/internal/compatibility/testdata"
	"go.uber.org/cadence/internal/compatibility/thrift"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

type ignoreFieldType struct{}

var ignoreField = ignoreFieldType{}

func assertAllFieldsSet(t *testing.T, obj interface{}, ignoredFields map[string]ignoreFieldType) {
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

	structType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := structType.Field(i)
		fieldName := fieldType.Name

		if ignoredFields != nil {
			if _, ignored := ignoredFields[fieldName]; ignored {
				continue
			}
		}

		if !field.CanInterface() {
			continue
		}

		// Ignore protobuf internal fields
		if strings.HasPrefix(fieldName, "XXX_") {
			continue
		}

		if isDefaultValue(field) {
			t.Errorf("Field %s.%s is set to default value, this may indicate missing mapper implementation", structType.Name(), fieldName)
		}
	}
}

func isDefaultValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Array:
		return v.Len() == 0
	case reflect.Struct:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	default:
		return false
	}
}

func TestActivityLocalDispatchInfo(t *testing.T) {
	for _, item := range []*apiv1.ActivityLocalDispatchInfo{nil, {}, &testdata.ActivityLocalDispatchInfo} {
		assert.Equal(t, item, proto.ActivityLocalDispatchInfo(thrift.ActivityLocalDispatchInfo(item)))
	}
	assertAllFieldsSet(t, &testdata.ActivityLocalDispatchInfo, nil)
}
func TestActivityTaskCancelRequestedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskCancelRequestedEventAttributes{nil, {}, &testdata.ActivityTaskCancelRequestedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskCancelRequestedEventAttributes(thrift.ActivityTaskCancelRequestedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ActivityTaskCancelRequestedEventAttributes, nil)
}
func TestActivityTaskCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskCanceledEventAttributes{nil, {}, &testdata.ActivityTaskCanceledEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskCanceledEventAttributes(thrift.ActivityTaskCanceledEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ActivityTaskCanceledEventAttributes, nil)
}
func TestActivityTaskCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskCompletedEventAttributes{nil, {}, &testdata.ActivityTaskCompletedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskCompletedEventAttributes(thrift.ActivityTaskCompletedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ActivityTaskCompletedEventAttributes, nil)
}
func TestActivityTaskFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskFailedEventAttributes{nil, {}, &testdata.ActivityTaskFailedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskFailedEventAttributes(thrift.ActivityTaskFailedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ActivityTaskFailedEventAttributes, nil)
}
func TestActivityTaskScheduledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskScheduledEventAttributes{nil, {}, &testdata.ActivityTaskScheduledEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskScheduledEventAttributes(thrift.ActivityTaskScheduledEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ActivityTaskScheduledEventAttributes, nil)
}
func TestActivityTaskStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskStartedEventAttributes{nil, {}, &testdata.ActivityTaskStartedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskStartedEventAttributes(thrift.ActivityTaskStartedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ActivityTaskStartedEventAttributes, nil)
}
func TestActivityTaskTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskTimedOutEventAttributes{nil, {}, &testdata.ActivityTaskTimedOutEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskTimedOutEventAttributes(thrift.ActivityTaskTimedOutEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ActivityTaskTimedOutEventAttributes, nil)
}
func TestActivityType(t *testing.T) {
	for _, item := range []*apiv1.ActivityType{nil, {}, &testdata.ActivityType} {
		assert.Equal(t, item, proto.ActivityType(thrift.ActivityType(item)))
	}
	assertAllFieldsSet(t, &testdata.ActivityType, nil)
}
func TestBadBinaries(t *testing.T) {
	for _, item := range []*apiv1.BadBinaries{nil, {}, &testdata.BadBinaries} {
		assert.Equal(t, item, proto.BadBinaries(thrift.BadBinaries(item)))
	}
	assertAllFieldsSet(t, &testdata.BadBinaries, nil)
}
func TestBadBinaryInfo(t *testing.T) {
	for _, item := range []*apiv1.BadBinaryInfo{nil, {}, &testdata.BadBinaryInfo} {
		assert.Equal(t, item, proto.BadBinaryInfo(thrift.BadBinaryInfo(item)))
	}
	assertAllFieldsSet(t, &testdata.BadBinaryInfo, nil)
}
func TestCancelTimerFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.CancelTimerFailedEventAttributes{nil, {}, &testdata.CancelTimerFailedEventAttributes} {
		assert.Equal(t, item, proto.CancelTimerFailedEventAttributes(thrift.CancelTimerFailedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.CancelTimerFailedEventAttributes, nil)
}
func TestChildWorkflowExecutionCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionCanceledEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionCanceledEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionCanceledEventAttributes(thrift.ChildWorkflowExecutionCanceledEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ChildWorkflowExecutionCanceledEventAttributes, nil)
}
func TestChildWorkflowExecutionCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionCompletedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionCompletedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionCompletedEventAttributes(thrift.ChildWorkflowExecutionCompletedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ChildWorkflowExecutionCompletedEventAttributes, nil)
}
func TestChildWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionFailedEventAttributes(thrift.ChildWorkflowExecutionFailedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ChildWorkflowExecutionFailedEventAttributes, nil)
}
func TestChildWorkflowExecutionStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionStartedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionStartedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionStartedEventAttributes(thrift.ChildWorkflowExecutionStartedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ChildWorkflowExecutionStartedEventAttributes, nil)
}
func TestChildWorkflowExecutionTerminatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionTerminatedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionTerminatedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionTerminatedEventAttributes(thrift.ChildWorkflowExecutionTerminatedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ChildWorkflowExecutionTerminatedEventAttributes, nil)
}
func TestChildWorkflowExecutionTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionTimedOutEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionTimedOutEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionTimedOutEventAttributes(thrift.ChildWorkflowExecutionTimedOutEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ChildWorkflowExecutionTimedOutEventAttributes, nil)
}
func TestClusterReplicationConfiguration(t *testing.T) {
	for _, item := range []*apiv1.ClusterReplicationConfiguration{nil, {}, &testdata.ClusterReplicationConfiguration} {
		assert.Equal(t, item, proto.ClusterReplicationConfiguration(thrift.ClusterReplicationConfiguration(item)))
	}
	assertAllFieldsSet(t, &testdata.ClusterReplicationConfiguration, nil)
}
func TestCountWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.CountWorkflowExecutionsRequest{nil, {}, &testdata.CountWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.CountWorkflowExecutionsRequest(thrift.CountWorkflowExecutionsRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.CountWorkflowExecutionsRequest, nil)
}
func TestCountWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.CountWorkflowExecutionsResponse{nil, {}, &testdata.CountWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.CountWorkflowExecutionsResponse(thrift.CountWorkflowExecutionsResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.CountWorkflowExecutionsResponse, nil)
}
func TestDataBlob(t *testing.T) {
	for _, item := range []*apiv1.DataBlob{nil, {}, &testdata.DataBlob} {
		assert.Equal(t, item, proto.DataBlob(thrift.DataBlob(item)))
	}
	assertAllFieldsSet(t, &testdata.DataBlob, nil)
}
func TestDecisionTaskCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskCompletedEventAttributes{nil, {}, &testdata.DecisionTaskCompletedEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskCompletedEventAttributes(thrift.DecisionTaskCompletedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.DecisionTaskCompletedEventAttributes, nil)
}
func TestDecisionTaskFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskFailedEventAttributes{nil, {}, &testdata.DecisionTaskFailedEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskFailedEventAttributes(thrift.DecisionTaskFailedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.DecisionTaskFailedEventAttributes, nil)
}
func TestDecisionTaskScheduledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskScheduledEventAttributes{nil, {}, &testdata.DecisionTaskScheduledEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskScheduledEventAttributes(thrift.DecisionTaskScheduledEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.DecisionTaskScheduledEventAttributes, nil)
}
func TestDecisionTaskStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskStartedEventAttributes{nil, {}, &testdata.DecisionTaskStartedEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskStartedEventAttributes(thrift.DecisionTaskStartedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.DecisionTaskStartedEventAttributes, nil)
}
func TestDecisionTaskTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskTimedOutEventAttributes{nil, {}, &testdata.DecisionTaskTimedOutEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskTimedOutEventAttributes(thrift.DecisionTaskTimedOutEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.DecisionTaskTimedOutEventAttributes, nil)
}
func TestDeprecateDomainRequest(t *testing.T) {
	for _, item := range []*apiv1.DeprecateDomainRequest{nil, {}, &testdata.DeprecateDomainRequest} {
		assert.Equal(t, item, proto.DeprecateDomainRequest(thrift.DeprecateDomainRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.DeprecateDomainRequest, nil)
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
	assertAllFieldsSet(t, &testdata.DescribeDomainRequest_ID, nil)
	assertAllFieldsSet(t, &testdata.DescribeDomainRequest_Name, nil)
}
func TestDescribeDomainResponse_Domain(t *testing.T) {
	for _, item := range []*apiv1.Domain{nil, &testdata.Domain} {
		assert.Equal(t, item, proto.DescribeDomainResponseDomain(thrift.DescribeDomainResponseDomain(item)))
	}
	assertAllFieldsSet(t, &testdata.Domain, nil)
}
func TestDescribeDomainResponse(t *testing.T) {
	for _, item := range []*apiv1.DescribeDomainResponse{nil, &testdata.DescribeDomainResponse} {
		assert.Equal(t, item, proto.DescribeDomainResponse(thrift.DescribeDomainResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.DescribeDomainResponse, nil)
}
func TestDescribeTaskListRequest(t *testing.T) {
	for _, item := range []*apiv1.DescribeTaskListRequest{nil, {}, &testdata.DescribeTaskListRequest} {
		assert.Equal(t, item, proto.DescribeTaskListRequest(thrift.DescribeTaskListRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.DescribeTaskListRequest, nil)
}
func TestDescribeTaskListResponse(t *testing.T) {
	for _, item := range []*apiv1.DescribeTaskListResponse{nil, {}, &testdata.DescribeTaskListResponse} {
		assert.Equal(t, item, proto.DescribeTaskListResponse(thrift.DescribeTaskListResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.DescribeTaskListResponse, nil)
}
func TestDescribeWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.DescribeWorkflowExecutionRequest{nil, {}, &testdata.DescribeWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.DescribeWorkflowExecutionRequest(thrift.DescribeWorkflowExecutionRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.DescribeWorkflowExecutionRequest, nil)
}
func TestDescribeWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.DescribeWorkflowExecutionResponse{nil, {}, &testdata.DescribeWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.DescribeWorkflowExecutionResponse(thrift.DescribeWorkflowExecutionResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.DescribeWorkflowExecutionResponse, nil)
}
func TestDiagnoseWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.DiagnoseWorkflowExecutionRequest{nil, {}, &testdata.DiagnoseWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.DiagnoseWorkflowExecutionRequest(thrift.DiagnoseWorkflowExecutionRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.DiagnoseWorkflowExecutionRequest, nil)
}
func TestDiagnoseWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.DiagnoseWorkflowExecutionResponse{nil, {}, &testdata.DiagnoseWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.DiagnoseWorkflowExecutionResponse(thrift.DiagnoseWorkflowExecutionResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.DiagnoseWorkflowExecutionResponse, nil)
}
func TestExternalWorkflowExecutionCancelRequestedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ExternalWorkflowExecutionCancelRequestedEventAttributes{nil, {}, &testdata.ExternalWorkflowExecutionCancelRequestedEventAttributes} {
		assert.Equal(t, item, proto.ExternalWorkflowExecutionCancelRequestedEventAttributes(thrift.ExternalWorkflowExecutionCancelRequestedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ExternalWorkflowExecutionCancelRequestedEventAttributes, nil)
}
func TestExternalWorkflowExecutionSignaledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ExternalWorkflowExecutionSignaledEventAttributes{nil, {}, &testdata.ExternalWorkflowExecutionSignaledEventAttributes} {
		assert.Equal(t, item, proto.ExternalWorkflowExecutionSignaledEventAttributes(thrift.ExternalWorkflowExecutionSignaledEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.ExternalWorkflowExecutionSignaledEventAttributes, nil)
}
func TestGetClusterInfoResponse(t *testing.T) {
	for _, item := range []*apiv1.GetClusterInfoResponse{nil, {}, &testdata.GetClusterInfoResponse} {
		assert.Equal(t, item, proto.GetClusterInfoResponse(thrift.GetClusterInfoResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.GetClusterInfoResponse, nil)
}
func TestGetSearchAttributesResponse(t *testing.T) {
	for _, item := range []*apiv1.GetSearchAttributesResponse{nil, {}, &testdata.GetSearchAttributesResponse} {
		assert.Equal(t, item, proto.GetSearchAttributesResponse(thrift.GetSearchAttributesResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.GetSearchAttributesResponse, nil)
}
func TestGetWorkflowExecutionHistoryRequest(t *testing.T) {
	for _, item := range []*apiv1.GetWorkflowExecutionHistoryRequest{nil, {}, &testdata.GetWorkflowExecutionHistoryRequest} {
		assert.Equal(t, item, proto.GetWorkflowExecutionHistoryRequest(thrift.GetWorkflowExecutionHistoryRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.GetWorkflowExecutionHistoryRequest, nil)
}
func TestGetWorkflowExecutionHistoryResponse(t *testing.T) {
	for _, item := range []*apiv1.GetWorkflowExecutionHistoryResponse{nil, {}, &testdata.GetWorkflowExecutionHistoryResponse} {
		assert.Equal(t, item, proto.GetWorkflowExecutionHistoryResponse(thrift.GetWorkflowExecutionHistoryResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.GetWorkflowExecutionHistoryResponse, nil)
}
func TestHeader(t *testing.T) {
	for _, item := range []*apiv1.Header{nil, {}, &testdata.Header} {
		assert.Equal(t, item, proto.Header(thrift.Header(item)))
	}
	assertAllFieldsSet(t, &testdata.Header, nil)
}
func TestHistory(t *testing.T) {
	for _, item := range []*apiv1.History{nil, {}, &testdata.History} {
		assert.Equal(t, item, proto.History(thrift.History(item)))
	}
	assertAllFieldsSet(t, &testdata.History, nil)
}
func TestListArchivedWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListArchivedWorkflowExecutionsRequest{nil, {}, &testdata.ListArchivedWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.ListArchivedWorkflowExecutionsRequest(thrift.ListArchivedWorkflowExecutionsRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.ListArchivedWorkflowExecutionsRequest, nil)
}
func TestListArchivedWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListArchivedWorkflowExecutionsResponse{nil, {}, &testdata.ListArchivedWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListArchivedWorkflowExecutionsResponse(thrift.ListArchivedWorkflowExecutionsResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.ListArchivedWorkflowExecutionsResponse, nil)
}
func TestListClosedWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListClosedWorkflowExecutionsResponse{nil, {}, &testdata.ListClosedWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListClosedWorkflowExecutionsResponse(thrift.ListClosedWorkflowExecutionsResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.ListClosedWorkflowExecutionsResponse, nil)
}
func TestListDomainsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListDomainsRequest{nil, {}, &testdata.ListDomainsRequest} {
		assert.Equal(t, item, proto.ListDomainsRequest(thrift.ListDomainsRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.ListDomainsRequest, nil)
}
func TestListDomainsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListDomainsResponse{nil, {}, &testdata.ListDomainsResponse} {
		assert.Equal(t, item, proto.ListDomainsResponse(thrift.ListDomainsResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.ListDomainsResponse, nil)
}
func TestListOpenWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListOpenWorkflowExecutionsResponse{nil, {}, &testdata.ListOpenWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListOpenWorkflowExecutionsResponse(thrift.ListOpenWorkflowExecutionsResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.ListOpenWorkflowExecutionsResponse, nil)
}
func TestListTaskListPartitionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListTaskListPartitionsRequest{nil, {}, &testdata.ListTaskListPartitionsRequest} {
		assert.Equal(t, item, proto.ListTaskListPartitionsRequest(thrift.ListTaskListPartitionsRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.ListTaskListPartitionsRequest, nil)
}
func TestListTaskListPartitionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListTaskListPartitionsResponse{nil, {}, &testdata.ListTaskListPartitionsResponse} {
		assert.Equal(t, item, proto.ListTaskListPartitionsResponse(thrift.ListTaskListPartitionsResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.ListTaskListPartitionsResponse, nil)
}
func TestListWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListWorkflowExecutionsRequest{nil, {}, &testdata.ListWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.ListWorkflowExecutionsRequest(thrift.ListWorkflowExecutionsRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.ListWorkflowExecutionsRequest, nil)
}
func TestListWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListWorkflowExecutionsResponse{nil, {}, &testdata.ListWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListWorkflowExecutionsResponse(thrift.ListWorkflowExecutionsResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.ListWorkflowExecutionsResponse, nil)
}
func TestMarkerRecordedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.MarkerRecordedEventAttributes{nil, {}, &testdata.MarkerRecordedEventAttributes} {
		assert.Equal(t, item, proto.MarkerRecordedEventAttributes(thrift.MarkerRecordedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.MarkerRecordedEventAttributes, nil)
}
func TestMemo(t *testing.T) {
	for _, item := range []*apiv1.Memo{nil, {}, &testdata.Memo} {
		assert.Equal(t, item, proto.Memo(thrift.Memo(item)))
	}
	assertAllFieldsSet(t, &testdata.Memo, nil)
}
func TestPendingActivityInfo(t *testing.T) {
	for _, item := range []*apiv1.PendingActivityInfo{nil, {}, &testdata.PendingActivityInfo} {
		assert.Equal(t, item, proto.PendingActivityInfo(thrift.PendingActivityInfo(item)))
	}
	assertAllFieldsSet(t, &testdata.PendingActivityInfo, nil)
}
func TestPendingChildExecutionInfo(t *testing.T) {
	for _, item := range []*apiv1.PendingChildExecutionInfo{nil, {}, &testdata.PendingChildExecutionInfo} {
		assert.Equal(t, item, proto.PendingChildExecutionInfo(thrift.PendingChildExecutionInfo(item)))
	}
	assertAllFieldsSet(t, &testdata.PendingChildExecutionInfo, nil)
}
func TestPendingDecisionInfo(t *testing.T) {
	for _, item := range []*apiv1.PendingDecisionInfo{nil, {}, &testdata.PendingDecisionInfo} {
		assert.Equal(t, item, proto.PendingDecisionInfo(thrift.PendingDecisionInfo(item)))
	}
	assertAllFieldsSet(t, &testdata.PendingDecisionInfo, nil)
}
func TestPollForActivityTaskRequest(t *testing.T) {
	for _, item := range []*apiv1.PollForActivityTaskRequest{nil, {}, &testdata.PollForActivityTaskRequest} {
		assert.Equal(t, item, proto.PollForActivityTaskRequest(thrift.PollForActivityTaskRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.PollForActivityTaskRequest, nil)
}
func TestPollForActivityTaskResponse(t *testing.T) {
	for _, item := range []*apiv1.PollForActivityTaskResponse{nil, {}, &testdata.PollForActivityTaskResponse} {
		assert.Equal(t, item, proto.PollForActivityTaskResponse(thrift.PollForActivityTaskResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.PollForActivityTaskResponse, nil)
}
func TestPollForDecisionTaskRequest(t *testing.T) {
	for _, item := range []*apiv1.PollForDecisionTaskRequest{nil, {}, &testdata.PollForDecisionTaskRequest} {
		assert.Equal(t, item, proto.PollForDecisionTaskRequest(thrift.PollForDecisionTaskRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.PollForDecisionTaskRequest, nil)
}
func TestPollForDecisionTaskResponse(t *testing.T) {
	for _, item := range []*apiv1.PollForDecisionTaskResponse{nil, {}, &testdata.PollForDecisionTaskResponse} {
		assert.Equal(t, item, proto.PollForDecisionTaskResponse(thrift.PollForDecisionTaskResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.PollForDecisionTaskResponse, nil)
}
func TestPollerInfo(t *testing.T) {
	for _, item := range []*apiv1.PollerInfo{nil, {}, &testdata.PollerInfo} {
		assert.Equal(t, item, proto.PollerInfo(thrift.PollerInfo(item)))
	}
	assertAllFieldsSet(t, &testdata.PollerInfo, nil)
}
func TestQueryRejected(t *testing.T) {
	for _, item := range []*apiv1.QueryRejected{nil, {}, &testdata.QueryRejected} {
		assert.Equal(t, item, proto.QueryRejected(thrift.QueryRejected(item)))
	}
	assertAllFieldsSet(t, &testdata.QueryRejected, nil)
}
func TestQueryWorkflowRequest(t *testing.T) {
	for _, item := range []*apiv1.QueryWorkflowRequest{nil, {}, &testdata.QueryWorkflowRequest} {
		assert.Equal(t, item, proto.QueryWorkflowRequest(thrift.QueryWorkflowRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.QueryWorkflowRequest, nil)
}
func TestQueryWorkflowResponse(t *testing.T) {
	for _, item := range []*apiv1.QueryWorkflowResponse{nil, {}, &testdata.QueryWorkflowResponse} {
		assert.Equal(t, item, proto.QueryWorkflowResponse(thrift.QueryWorkflowResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.QueryWorkflowResponse, nil)
}
func TestRecordActivityTaskHeartbeatByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatByIDRequest{nil, {}, &testdata.RecordActivityTaskHeartbeatByIDRequest} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatByIDRequest(thrift.RecordActivityTaskHeartbeatByIDRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.RecordActivityTaskHeartbeatByIDRequest, nil)
}
func TestRecordActivityTaskHeartbeatByIDResponse(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatByIDResponse{nil, {}, &testdata.RecordActivityTaskHeartbeatByIDResponse} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatByIDResponse(thrift.RecordActivityTaskHeartbeatByIDResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.RecordActivityTaskHeartbeatByIDResponse, nil)
}
func TestRecordActivityTaskHeartbeatRequest(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatRequest{nil, {}, &testdata.RecordActivityTaskHeartbeatRequest} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatRequest(thrift.RecordActivityTaskHeartbeatRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.RecordActivityTaskHeartbeatRequest, nil)
}
func TestRecordActivityTaskHeartbeatResponse(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatResponse{nil, {}, &testdata.RecordActivityTaskHeartbeatResponse} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatResponse(thrift.RecordActivityTaskHeartbeatResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.RecordActivityTaskHeartbeatResponse, nil)
}
func TestRegisterDomainRequest(t *testing.T) {
	for _, item := range []*apiv1.RegisterDomainRequest{nil, &testdata.RegisterDomainRequest} {
		assert.Equal(t, item, proto.RegisterDomainRequest(thrift.RegisterDomainRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.RegisterDomainRequest, nil)
}
func TestRequestCancelActivityTaskFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelActivityTaskFailedEventAttributes{nil, {}, &testdata.RequestCancelActivityTaskFailedEventAttributes} {
		assert.Equal(t, item, proto.RequestCancelActivityTaskFailedEventAttributes(thrift.RequestCancelActivityTaskFailedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.RequestCancelActivityTaskFailedEventAttributes, nil)
}
func TestRequestCancelExternalWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelExternalWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.RequestCancelExternalWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.RequestCancelExternalWorkflowExecutionFailedEventAttributes(thrift.RequestCancelExternalWorkflowExecutionFailedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.RequestCancelExternalWorkflowExecutionFailedEventAttributes, nil)
}
func TestRequestCancelExternalWorkflowExecutionInitiatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{nil, {}, &testdata.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes} {
		assert.Equal(t, item, proto.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes(thrift.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes, nil)
}
func TestRequestCancelWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelWorkflowExecutionRequest{nil, {}, &testdata.RequestCancelWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.RequestCancelWorkflowExecutionRequest(thrift.RequestCancelWorkflowExecutionRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.RequestCancelWorkflowExecutionRequest, nil)
}
func TestResetPointInfo(t *testing.T) {
	for _, item := range []*apiv1.ResetPointInfo{nil, {}, &testdata.ResetPointInfo} {
		assert.Equal(t, item, proto.ResetPointInfo(thrift.ResetPointInfo(item)))
	}
	assertAllFieldsSet(t, &testdata.ResetPointInfo, nil)
}
func TestResetPoints(t *testing.T) {
	for _, item := range []*apiv1.ResetPoints{nil, {}, &testdata.ResetPoints} {
		assert.Equal(t, item, proto.ResetPoints(thrift.ResetPoints(item)))
	}
	assertAllFieldsSet(t, &testdata.ResetPoints, nil)
}
func TestResetStickyTaskListRequest(t *testing.T) {
	for _, item := range []*apiv1.ResetStickyTaskListRequest{nil, {}, &testdata.ResetStickyTaskListRequest} {
		assert.Equal(t, item, proto.ResetStickyTaskListRequest(thrift.ResetStickyTaskListRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.ResetStickyTaskListRequest, nil)
}
func TestResetWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.ResetWorkflowExecutionRequest{nil, {}, &testdata.ResetWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.ResetWorkflowExecutionRequest(thrift.ResetWorkflowExecutionRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.ResetWorkflowExecutionRequest, nil)
}
func TestResetWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.ResetWorkflowExecutionResponse{nil, {}, &testdata.ResetWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.ResetWorkflowExecutionResponse(thrift.ResetWorkflowExecutionResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.ResetWorkflowExecutionResponse, nil)
}
func TestRespondActivityTaskCanceledByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCanceledByIDRequest{nil, {}, &testdata.RespondActivityTaskCanceledByIDRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCanceledByIDRequest(thrift.RespondActivityTaskCanceledByIDRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.RespondActivityTaskCanceledByIDRequest, nil)
}
func TestRespondActivityTaskCanceledRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCanceledRequest{nil, {}, &testdata.RespondActivityTaskCanceledRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCanceledRequest(thrift.RespondActivityTaskCanceledRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.RespondActivityTaskCanceledRequest, nil)
}
func TestRespondActivityTaskCompletedByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCompletedByIDRequest{nil, {}, &testdata.RespondActivityTaskCompletedByIDRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCompletedByIDRequest(thrift.RespondActivityTaskCompletedByIDRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.RespondActivityTaskCompletedByIDRequest, nil)
}
func TestRespondActivityTaskCompletedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCompletedRequest{nil, {}, &testdata.RespondActivityTaskCompletedRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCompletedRequest(thrift.RespondActivityTaskCompletedRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.RespondActivityTaskCompletedRequest, nil)
}
func TestRespondActivityTaskFailedByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskFailedByIDRequest{nil, {}, &testdata.RespondActivityTaskFailedByIDRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskFailedByIDRequest(thrift.RespondActivityTaskFailedByIDRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.RespondActivityTaskFailedByIDRequest, nil)
}
func TestRespondActivityTaskFailedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskFailedRequest{nil, {}, &testdata.RespondActivityTaskFailedRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskFailedRequest(thrift.RespondActivityTaskFailedRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.RespondActivityTaskFailedRequest, nil)
}
func TestRespondDecisionTaskCompletedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondDecisionTaskCompletedRequest{nil, {}, &testdata.RespondDecisionTaskCompletedRequest} {
		assert.Equal(t, item, proto.RespondDecisionTaskCompletedRequest(thrift.RespondDecisionTaskCompletedRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.RespondDecisionTaskCompletedRequest, nil)
}
func TestRespondDecisionTaskCompletedResponse(t *testing.T) {
	for _, item := range []*apiv1.RespondDecisionTaskCompletedResponse{nil, {}, &testdata.RespondDecisionTaskCompletedResponse} {
		assert.Equal(t, item, proto.RespondDecisionTaskCompletedResponse(thrift.RespondDecisionTaskCompletedResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.RespondDecisionTaskCompletedResponse, nil)
}
func TestRespondDecisionTaskFailedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondDecisionTaskFailedRequest{nil, {}, &testdata.RespondDecisionTaskFailedRequest} {
		assert.Equal(t, item, proto.RespondDecisionTaskFailedRequest(thrift.RespondDecisionTaskFailedRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.RespondDecisionTaskFailedRequest, nil)
}
func TestRespondQueryTaskCompletedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondQueryTaskCompletedRequest{nil, {Result: &apiv1.WorkflowQueryResult{}}, &testdata.RespondQueryTaskCompletedRequest} {
		assert.Equal(t, item, proto.RespondQueryTaskCompletedRequest(thrift.RespondQueryTaskCompletedRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.RespondQueryTaskCompletedRequest, nil)
}
func TestRetryPolicy(t *testing.T) {
	for _, item := range []*apiv1.RetryPolicy{nil, {}, &testdata.RetryPolicy} {
		assert.Equal(t, item, proto.RetryPolicy(thrift.RetryPolicy(item)))
	}
	assertAllFieldsSet(t, &testdata.RetryPolicy, nil)
}
func TestScanWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ScanWorkflowExecutionsRequest{nil, {}, &testdata.ScanWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.ScanWorkflowExecutionsRequest(thrift.ScanWorkflowExecutionsRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.ScanWorkflowExecutionsRequest, nil)
}
func TestScanWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ScanWorkflowExecutionsResponse{nil, {}, &testdata.ScanWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ScanWorkflowExecutionsResponse(thrift.ScanWorkflowExecutionsResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.ScanWorkflowExecutionsResponse, nil)
}
func TestSearchAttributes(t *testing.T) {
	for _, item := range []*apiv1.SearchAttributes{nil, {}, &testdata.SearchAttributes} {
		assert.Equal(t, item, proto.SearchAttributes(thrift.SearchAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.SearchAttributes, nil)
}
func TestSignalExternalWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.SignalExternalWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.SignalExternalWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.SignalExternalWorkflowExecutionFailedEventAttributes(thrift.SignalExternalWorkflowExecutionFailedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.SignalExternalWorkflowExecutionFailedEventAttributes, nil)
}
func TestSignalExternalWorkflowExecutionInitiatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.SignalExternalWorkflowExecutionInitiatedEventAttributes{nil, {}, &testdata.SignalExternalWorkflowExecutionInitiatedEventAttributes} {
		assert.Equal(t, item, proto.SignalExternalWorkflowExecutionInitiatedEventAttributes(thrift.SignalExternalWorkflowExecutionInitiatedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.SignalExternalWorkflowExecutionInitiatedEventAttributes, nil)
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
	assertAllFieldsSet(t, &testdata.SignalWithStartWorkflowExecutionRequest, nil)
	assertAllFieldsSet(t, &testdata.SignalWithStartWorkflowExecutionRequestWithCronAndActiveClusterSelectionPolicy1, nil)
	assertAllFieldsSet(t, &testdata.SignalWithStartWorkflowExecutionRequestWithCronAndActiveClusterSelectionPolicy2, nil)
}
func TestSignalWithStartWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.SignalWithStartWorkflowExecutionResponse{nil, {}, &testdata.SignalWithStartWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.SignalWithStartWorkflowExecutionResponse(thrift.SignalWithStartWorkflowExecutionResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.SignalWithStartWorkflowExecutionResponse, nil)
}
func TestSignalWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.SignalWorkflowExecutionRequest{nil, {}, &testdata.SignalWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.SignalWorkflowExecutionRequest(thrift.SignalWorkflowExecutionRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.SignalWorkflowExecutionRequest, nil)
}
func TestStartChildWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.StartChildWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.StartChildWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.StartChildWorkflowExecutionFailedEventAttributes(thrift.StartChildWorkflowExecutionFailedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.StartChildWorkflowExecutionFailedEventAttributes, nil)
}
func TestStartChildWorkflowExecutionInitiatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.StartChildWorkflowExecutionInitiatedEventAttributes{nil, {}, &testdata.StartChildWorkflowExecutionInitiatedEventAttributes} {
		assert.Equal(t, item, proto.StartChildWorkflowExecutionInitiatedEventAttributes(thrift.StartChildWorkflowExecutionInitiatedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.StartChildWorkflowExecutionInitiatedEventAttributes, nil)
}
func TestStartTimeFilter(t *testing.T) {
	for _, item := range []*apiv1.StartTimeFilter{nil, {}, &testdata.StartTimeFilter} {
		assert.Equal(t, item, proto.StartTimeFilter(thrift.StartTimeFilter(item)))
	}
	assertAllFieldsSet(t, &testdata.StartTimeFilter, nil)
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
	assertAllFieldsSet(t, &testdata.StartWorkflowExecutionRequest, nil)
	assertAllFieldsSet(t, &testdata.StartWorkflowExecutionRequestWithCronAndActiveClusterSelectionPolicy1, nil)
	assertAllFieldsSet(t, &testdata.StartWorkflowExecutionRequestWithCronAndActiveClusterSelectionPolicy2, nil)
}
func TestStartWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.StartWorkflowExecutionResponse{nil, {}, &testdata.StartWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.StartWorkflowExecutionResponse(thrift.StartWorkflowExecutionResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.StartWorkflowExecutionResponse, nil)
}
func TestStatusFilter(t *testing.T) {
	for _, item := range []*apiv1.StatusFilter{nil, &testdata.StatusFilter} {
		assert.Equal(t, item, proto.StatusFilter(thrift.StatusFilter(item)))
	}
	assertAllFieldsSet(t, &testdata.StatusFilter, nil)
}
func TestStickyExecutionAttributes(t *testing.T) {
	for _, item := range []*apiv1.StickyExecutionAttributes{nil, {}, &testdata.StickyExecutionAttributes} {
		assert.Equal(t, item, proto.StickyExecutionAttributes(thrift.StickyExecutionAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.StickyExecutionAttributes, nil)
}
func TestSupportedClientVersions(t *testing.T) {
	for _, item := range []*apiv1.SupportedClientVersions{nil, {}, &testdata.SupportedClientVersions} {
		assert.Equal(t, item, proto.SupportedClientVersions(thrift.SupportedClientVersions(item)))
	}
	assertAllFieldsSet(t, &testdata.SupportedClientVersions, nil)
}
func TestTaskIDBlock(t *testing.T) {
	for _, item := range []*apiv1.TaskIDBlock{nil, {}, &testdata.TaskIDBlock} {
		assert.Equal(t, item, proto.TaskIDBlock(thrift.TaskIDBlock(item)))
	}
	assertAllFieldsSet(t, &testdata.TaskIDBlock, nil)
}
func TestTaskList(t *testing.T) {
	for _, item := range []*apiv1.TaskList{nil, {}, &testdata.TaskList} {
		assert.Equal(t, item, proto.TaskList(thrift.TaskList(item)))
	}
	assertAllFieldsSet(t, &testdata.TaskList, nil)
}
func TestTaskListMetadata(t *testing.T) {
	for _, item := range []*apiv1.TaskListMetadata{nil, {}, &testdata.TaskListMetadata} {
		assert.Equal(t, item, proto.TaskListMetadata(thrift.TaskListMetadata(item)))
	}
	assertAllFieldsSet(t, &testdata.TaskListMetadata, nil)
}
func TestTaskListPartitionMetadata(t *testing.T) {
	for _, item := range []*apiv1.TaskListPartitionMetadata{nil, {}, &testdata.TaskListPartitionMetadata} {
		assert.Equal(t, item, proto.TaskListPartitionMetadata(thrift.TaskListPartitionMetadata(item)))
	}
	assertAllFieldsSet(t, &testdata.TaskListPartitionMetadata, nil)
}
func TestTaskListStatus(t *testing.T) {
	for _, item := range []*apiv1.TaskListStatus{nil, {}, &testdata.TaskListStatus} {
		assert.Equal(t, item, proto.TaskListStatus(thrift.TaskListStatus(item)))
	}
	assertAllFieldsSet(t, &testdata.TaskListStatus, nil)
}
func TestTerminateWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.TerminateWorkflowExecutionRequest{nil, {}, &testdata.TerminateWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.TerminateWorkflowExecutionRequest(thrift.TerminateWorkflowExecutionRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.TerminateWorkflowExecutionRequest, nil)
}
func TestTimerCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.TimerCanceledEventAttributes{nil, {}, &testdata.TimerCanceledEventAttributes} {
		assert.Equal(t, item, proto.TimerCanceledEventAttributes(thrift.TimerCanceledEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.TimerCanceledEventAttributes, nil)
}
func TestTimerFiredEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.TimerFiredEventAttributes{nil, {}, &testdata.TimerFiredEventAttributes} {
		assert.Equal(t, item, proto.TimerFiredEventAttributes(thrift.TimerFiredEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.TimerFiredEventAttributes, nil)
}
func TestTimerStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.TimerStartedEventAttributes{nil, {}, &testdata.TimerStartedEventAttributes} {
		assert.Equal(t, item, proto.TimerStartedEventAttributes(thrift.TimerStartedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.TimerStartedEventAttributes, nil)
}
func TestUpdateDomainRequest(t *testing.T) {
	for _, item := range []*apiv1.UpdateDomainRequest{nil, {UpdateMask: &gogo.FieldMask{}}, &testdata.UpdateDomainRequest} {
		assert.Equal(t, item, proto.UpdateDomainRequest(thrift.UpdateDomainRequest(item)))
	}
	assertAllFieldsSet(t, &testdata.UpdateDomainRequest, nil)
}
func TestUpdateDomainResponse(t *testing.T) {
	for _, item := range []*apiv1.UpdateDomainResponse{nil, &testdata.UpdateDomainResponse} {
		assert.Equal(t, item, proto.UpdateDomainResponse(thrift.UpdateDomainResponse(item)))
	}
	assertAllFieldsSet(t, &testdata.UpdateDomainResponse, nil)
}
func TestUpsertWorkflowSearchAttributesEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.UpsertWorkflowSearchAttributesEventAttributes{nil, {}, &testdata.UpsertWorkflowSearchAttributesEventAttributes} {
		assert.Equal(t, item, proto.UpsertWorkflowSearchAttributesEventAttributes(thrift.UpsertWorkflowSearchAttributesEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.UpsertWorkflowSearchAttributesEventAttributes, nil)
}
func TestWorkerVersionInfo(t *testing.T) {
	for _, item := range []*apiv1.WorkerVersionInfo{nil, {}, &testdata.WorkerVersionInfo} {
		assert.Equal(t, item, proto.WorkerVersionInfo(thrift.WorkerVersionInfo(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkerVersionInfo, nil)
}
func TestWorkflowExecution(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecution{nil, {}, &testdata.WorkflowExecution} {
		assert.Equal(t, item, proto.WorkflowExecution(thrift.WorkflowExecution(item)))
	}
	assert.Empty(t, thrift.WorkflowID(nil))
	assert.Empty(t, thrift.RunID(nil))
	assertAllFieldsSet(t, &testdata.WorkflowExecution, nil)
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
	assertAllFieldsSet(t, &testdata.WorkflowExecutionCancelRequestedEventAttributes, nil)
}
func TestWorkflowExecutionCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionCanceledEventAttributes{nil, {}, &testdata.WorkflowExecutionCanceledEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionCanceledEventAttributes(thrift.WorkflowExecutionCanceledEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkflowExecutionCanceledEventAttributes, nil)
}
func TestWorkflowExecutionCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionCompletedEventAttributes{nil, {}, &testdata.WorkflowExecutionCompletedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionCompletedEventAttributes(thrift.WorkflowExecutionCompletedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkflowExecutionCompletedEventAttributes, nil)
}
func TestWorkflowExecutionConfiguration(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionConfiguration{nil, {}, &testdata.WorkflowExecutionConfiguration} {
		assert.Equal(t, item, proto.WorkflowExecutionConfiguration(thrift.WorkflowExecutionConfiguration(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkflowExecutionConfiguration, nil)
}
func TestWorkflowExecutionContinuedAsNewEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionContinuedAsNewEventAttributes{nil, {}, &testdata.WorkflowExecutionContinuedAsNewEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionContinuedAsNewEventAttributes(thrift.WorkflowExecutionContinuedAsNewEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkflowExecutionContinuedAsNewEventAttributes, nil)
}
func TestWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionFailedEventAttributes{nil, {}, &testdata.WorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionFailedEventAttributes(thrift.WorkflowExecutionFailedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkflowExecutionFailedEventAttributes, nil)
}
func TestWorkflowExecutionFilter(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionFilter{nil, {}, &testdata.WorkflowExecutionFilter} {
		assert.Equal(t, item, proto.WorkflowExecutionFilter(thrift.WorkflowExecutionFilter(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkflowExecutionFilter, nil)
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
	assertAllFieldsSet(t, &testdata.WorkflowExecutionInfo, nil)
}
func TestWorkflowExecutionSignaledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionSignaledEventAttributes{nil, {}, &testdata.WorkflowExecutionSignaledEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionSignaledEventAttributes(thrift.WorkflowExecutionSignaledEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkflowExecutionSignaledEventAttributes, nil)
}
func TestWorkflowExecutionStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionStartedEventAttributes{nil, {}, &testdata.WorkflowExecutionStartedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionStartedEventAttributes(thrift.WorkflowExecutionStartedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkflowExecutionStartedEventAttributes, nil)
}
func TestWorkflowExecutionTerminatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionTerminatedEventAttributes{nil, {}, &testdata.WorkflowExecutionTerminatedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionTerminatedEventAttributes(thrift.WorkflowExecutionTerminatedEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkflowExecutionTerminatedEventAttributes, nil)
}
func TestWorkflowExecutionTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionTimedOutEventAttributes{nil, {}, &testdata.WorkflowExecutionTimedOutEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionTimedOutEventAttributes(thrift.WorkflowExecutionTimedOutEventAttributes(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkflowExecutionTimedOutEventAttributes, nil)
}
func TestWorkflowQuery(t *testing.T) {
	for _, item := range []*apiv1.WorkflowQuery{nil, {}, &testdata.WorkflowQuery} {
		assert.Equal(t, item, proto.WorkflowQuery(thrift.WorkflowQuery(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkflowQuery, nil)
}
func TestWorkflowQueryResult(t *testing.T) {
	for _, item := range []*apiv1.WorkflowQueryResult{nil, {}, &testdata.WorkflowQueryResult} {
		assert.Equal(t, item, proto.WorkflowQueryResult(thrift.WorkflowQueryResult(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkflowQueryResult, nil)
}
func TestWorkflowType(t *testing.T) {
	for _, item := range []*apiv1.WorkflowType{nil, {}, &testdata.WorkflowType} {
		assert.Equal(t, item, proto.WorkflowType(thrift.WorkflowType(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkflowType, nil)
}
func TestWorkflowTypeFilter(t *testing.T) {
	for _, item := range []*apiv1.WorkflowTypeFilter{nil, {}, &testdata.WorkflowTypeFilter} {
		assert.Equal(t, item, proto.WorkflowTypeFilter(thrift.WorkflowTypeFilter(item)))
	}
	assertAllFieldsSet(t, &testdata.WorkflowTypeFilter, nil)
}
func TestDataBlobArray(t *testing.T) {
	for _, item := range [][]*apiv1.DataBlob{nil, {}, testdata.DataBlobArray} {
		assert.Equal(t, item, proto.DataBlobArray(thrift.DataBlobArray(item)))
	}
	for _, testObj := range testdata.DataBlobArray {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestHistoryEventArray(t *testing.T) {
	for _, item := range [][]*apiv1.HistoryEvent{nil, {}, testdata.HistoryEventArray} {
		assert.Equal(t, item, proto.HistoryEventArray(thrift.HistoryEventArray(item)))
	}
	for _, testObj := range testdata.HistoryEventArray {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestTaskListPartitionMetadataArray(t *testing.T) {
	for _, item := range [][]*apiv1.TaskListPartitionMetadata{nil, {}, testdata.TaskListPartitionMetadataArray} {
		assert.Equal(t, item, proto.TaskListPartitionMetadataArray(thrift.TaskListPartitionMetadataArray(item)))
	}
	for _, testObj := range testdata.TaskListPartitionMetadataArray {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestDecisionArray(t *testing.T) {
	for _, item := range [][]*apiv1.Decision{nil, {}, testdata.DecisionArray} {
		assert.Equal(t, item, proto.DecisionArray(thrift.DecisionArray(item)))
	}
	for _, testObj := range testdata.DecisionArray {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestPollerInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.PollerInfo{nil, {}, testdata.PollerInfoArray} {
		assert.Equal(t, item, proto.PollerInfoArray(thrift.PollerInfoArray(item)))
	}
	for _, testObj := range testdata.PollerInfoArray {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestPendingChildExecutionInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.PendingChildExecutionInfo{nil, {}, testdata.PendingChildExecutionInfoArray} {
		assert.Equal(t, item, proto.PendingChildExecutionInfoArray(thrift.PendingChildExecutionInfoArray(item)))
	}
	for _, testObj := range testdata.PendingChildExecutionInfoArray {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestWorkflowExecutionInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.WorkflowExecutionInfo{nil, {}, testdata.WorkflowExecutionInfoArray} {
		assert.Equal(t, item, proto.WorkflowExecutionInfoArray(thrift.WorkflowExecutionInfoArray(item)))
	}
	for _, testObj := range testdata.WorkflowExecutionInfoArray {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestDescribeDomainResponseArray(t *testing.T) {
	for _, item := range [][]*apiv1.Domain{nil, {}, testdata.DomainArray} {
		assert.Equal(t, item, proto.DescribeDomainResponseArray(thrift.DescribeDomainResponseArray(item)))
	}
	for _, testObj := range testdata.DomainArray {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestResetPointInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.ResetPointInfo{nil, {}, testdata.ResetPointInfoArray} {
		assert.Equal(t, item, proto.ResetPointInfoArray(thrift.ResetPointInfoArray(item)))
	}
	for _, testObj := range testdata.ResetPointInfoArray {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestPendingActivityInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.PendingActivityInfo{nil, {}, testdata.PendingActivityInfoArray} {
		assert.Equal(t, item, proto.PendingActivityInfoArray(thrift.PendingActivityInfoArray(item)))
	}
	for _, testObj := range testdata.PendingActivityInfoArray {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestClusterReplicationConfigurationArray(t *testing.T) {
	for _, item := range [][]*apiv1.ClusterReplicationConfiguration{nil, {}, testdata.ClusterReplicationConfigurationArray} {
		assert.Equal(t, item, proto.ClusterReplicationConfigurationArray(thrift.ClusterReplicationConfigurationArray(item)))
	}
	for _, testObj := range testdata.ClusterReplicationConfigurationArray {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestActivityLocalDispatchInfoMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.ActivityLocalDispatchInfo{nil, {}, testdata.ActivityLocalDispatchInfoMap} {
		assert.Equal(t, item, proto.ActivityLocalDispatchInfoMap(thrift.ActivityLocalDispatchInfoMap(item)))
	}
	for _, testObj := range testdata.ActivityLocalDispatchInfoMap {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestBadBinaryInfoMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.BadBinaryInfo{nil, {}, testdata.BadBinaryInfoMap} {
		assert.Equal(t, item, proto.BadBinaryInfoMap(thrift.BadBinaryInfoMap(item)))
	}
	for _, testObj := range testdata.BadBinaryInfoMap {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestIndexedValueTypeMap(t *testing.T) {
	for _, item := range []map[string]apiv1.IndexedValueType{nil, {}, testdata.IndexedValueTypeMap} {
		assert.Equal(t, item, proto.IndexedValueTypeMap(thrift.IndexedValueTypeMap(item)))
	}
	// Note: IndexedValueType is an enum, not a struct, so no assertAllFieldsSet needed
}
func TestWorkflowQueryMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.WorkflowQuery{nil, {}, testdata.WorkflowQueryMap} {
		assert.Equal(t, item, proto.WorkflowQueryMap(thrift.WorkflowQueryMap(item)))
	}
	for _, testObj := range testdata.WorkflowQueryMap {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestWorkflowQueryResultMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.WorkflowQueryResult{nil, {}, testdata.WorkflowQueryResultMap} {
		assert.Equal(t, item, proto.WorkflowQueryResultMap(thrift.WorkflowQueryResultMap(item)))
	}
	for _, testObj := range testdata.WorkflowQueryResultMap {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
}
func TestPayload(t *testing.T) {
	for _, item := range []*apiv1.Payload{nil, &testdata.Payload1} {
		assert.Equal(t, item, proto.Payload(thrift.Payload(item)))
	}

	assert.Equal(t, &apiv1.Payload{Data: []byte{}}, proto.Payload(thrift.Payload(&apiv1.Payload{})))
	assertAllFieldsSet(t, &testdata.Payload1, nil)
}
func TestPayloadMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.Payload{nil, {}, testdata.PayloadMap} {
		assert.Equal(t, item, proto.PayloadMap(thrift.PayloadMap(item)))
	}
	for _, testObj := range testdata.PayloadMap {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
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

	// Assert all fields are set for non-nil history events
	for _, testObj := range historyEvents {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
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

	// Assert all fields are set for non-nil decisions
	for _, testObj := range decisions {
		if testObj != nil {
			assertAllFieldsSet(t, testObj, nil)
		}
	}
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
	assertAllFieldsSet(t, &testdata.ListClosedWorkflowExecutionsRequest_ExecutionFilter, nil)
	assertAllFieldsSet(t, &testdata.ListClosedWorkflowExecutionsRequest_StatusFilter, nil)
	assertAllFieldsSet(t, &testdata.ListClosedWorkflowExecutionsRequest_TypeFilter, nil)
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
	assertAllFieldsSet(t, &testdata.ListOpenWorkflowExecutionsRequest_ExecutionFilter, nil)
	assertAllFieldsSet(t, &testdata.ListOpenWorkflowExecutionsRequest_TypeFilter, nil)
}
