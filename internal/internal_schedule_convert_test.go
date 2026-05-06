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

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

// ── Enum converters ───────────────────────────────────────────────────────────

func TestScheduleOverlapPolicy_EnumRoundTrip(t *testing.T) {
	cases := []struct {
		sdk   ScheduleOverlapPolicy
		proto apiv1.ScheduleOverlapPolicy
	}{
		{ScheduleOverlapPolicyUnspecified, apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_INVALID},
		{ScheduleOverlapPolicySkipNew, apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_SKIP_NEW},
		{ScheduleOverlapPolicyBuffer, apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_BUFFER},
		{ScheduleOverlapPolicyConcurrent, apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_CONCURRENT},
		{ScheduleOverlapPolicyCancelPrevious, apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_CANCEL_PREVIOUS},
		{ScheduleOverlapPolicyTerminatePrevious, apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_TERMINATE_PREVIOUS},
	}
	for _, c := range cases {
		assert.Equal(t, c.proto, c.sdk.toProto(), "sdk→proto for %v", c.sdk)
		assert.Equal(t, c.sdk, scheduleOverlapPolicyFromProto(c.proto), "proto→sdk for %v", c.proto)
	}
}

func TestScheduleCatchUpPolicy_EnumRoundTrip(t *testing.T) {
	cases := []struct {
		sdk   ScheduleCatchUpPolicy
		proto apiv1.ScheduleCatchUpPolicy
	}{
		{ScheduleCatchUpPolicyUnspecified, apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_INVALID},
		{ScheduleCatchUpPolicySkip, apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_SKIP},
		{ScheduleCatchUpPolicyOne, apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_ONE},
		{ScheduleCatchUpPolicyAll, apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_ALL},
	}
	for _, c := range cases {
		assert.Equal(t, c.proto, c.sdk.toProto(), "sdk→proto for %v", c.sdk)
		assert.Equal(t, c.sdk, scheduleCatchUpPolicyFromProto(c.proto), "proto→sdk for %v", c.proto)
	}
}

// ── Converter unit tests ──────────────────────────────────────────────────────

func TestFromProtoNilGuards(t *testing.T) {
	assert.Nil(t, scheduleStartWorkflowActionFromProto(nil))
	assert.Nil(t, scheduleActionFromProto(nil))
	assert.Nil(t, schedulePoliciesFromProto(nil))
	assert.Nil(t, schedulePauseInfoFromProto(nil))
	assert.Nil(t, scheduleStateFromProto(nil))
	assert.Nil(t, backfillInfoFromProto(nil))
	assert.Nil(t, scheduleInfoFromProto(nil))
	assert.Nil(t, scheduleRetryPolicyFromProto(nil))
	assert.Nil(t, scheduleListEntryFromProto(nil))
	assert.Nil(t, describeScheduleResponseFromProto(nil))
	assert.Nil(t, listSchedulesResponseFromProto(nil))
}

func TestToProtoNilGuards(t *testing.T) {
	dc := getDefaultDataConverter()
	assert.Nil(t, scheduleSpecToProto(nil))
	assert.Nil(t, schedulePoliciesToProto(nil))
	assert.Nil(t, scheduleRetryPolicyToProto(nil))
	result, err := scheduleStartWorkflowActionToProto(nil, dc)
	require.NoError(t, err)
	assert.Nil(t, result)
	protoAction, err := scheduleActionToProto(nil, dc)
	require.NoError(t, err)
	assert.Nil(t, protoAction)
}

func TestScheduleMemoConverter(t *testing.T) {
	dc := getDefaultDataConverter()
	t.Run("to proto - nil input returns nil", func(t *testing.T) {
		result, err := scheduleMemoToProto(nil, dc)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
	t.Run("to proto - non-empty memo is encoded", func(t *testing.T) {
		result, err := scheduleMemoToProto(map[string]interface{}{"key": "value"}, dc)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Contains(t, result.Fields, "key")
	})
	t.Run("to proto - encoding error", func(t *testing.T) {
		ch := make(chan int)
		_, err := scheduleMemoToProto(map[string]interface{}{"key": ch}, dc)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "encode memo field")
	})
	t.Run("from proto - nil returns nil", func(t *testing.T) {
		assert.Nil(t, scheduleMemoFromProto(nil))
	})
	t.Run("from proto - non-nil memo is decoded", func(t *testing.T) {
		memo := &apiv1.Memo{Fields: map[string]*apiv1.Payload{"k": {Data: []byte("v")}}}
		result := scheduleMemoFromProto(memo)
		require.NotNil(t, result)
		assert.Equal(t, []byte("v"), result["k"])
	})
}

func TestScheduleSearchAttributesConverter(t *testing.T) {
	t.Run("to proto - nil input returns nil", func(t *testing.T) {
		result, err := scheduleSearchAttributesToProto(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})
	t.Run("to proto - non-empty attrs are encoded", func(t *testing.T) {
		result, err := scheduleSearchAttributesToProto(map[string]interface{}{"attr": "val"})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Contains(t, result.IndexedFields, "attr")
	})
	t.Run("to proto - encoding error", func(t *testing.T) {
		ch := make(chan int)
		_, err := scheduleSearchAttributesToProto(map[string]interface{}{"key": ch})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "encode search attribute")
	})
	t.Run("from proto - nil returns nil", func(t *testing.T) {
		assert.Nil(t, scheduleSearchAttributesFromProto(nil))
	})
	t.Run("from proto - non-nil attrs are decoded", func(t *testing.T) {
		sa := &apiv1.SearchAttributes{IndexedFields: map[string]*apiv1.Payload{"attr": {Data: []byte(`"val"`)}}}
		result := scheduleSearchAttributesFromProto(sa)
		require.NotNil(t, result)
		assert.Equal(t, []byte(`"val"`), result["attr"])
	})
}
