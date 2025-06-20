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

package internal

import (
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/cadence/internal/common/testlogger"

	"github.com/opentracing/opentracing-go"
	"github.com/uber-go/tally"

	"go.uber.org/cadence/internal/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	s "go.uber.org/cadence/.gen/go/shared"
)

func TestReplayAwareLogger(t *testing.T) {
	t.Parallel()
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core, zap.Development())

	isReplay, enableLoggingInReplay := false, false
	logger = logger.WithOptions(zap.WrapCore(wrapLogger(&isReplay, &enableLoggingInReplay)))

	logger.Info("normal info")

	isReplay = true
	logger.Info("replay info") // this log should be suppressed

	isReplay, enableLoggingInReplay = false, true
	logger.Info("normal2 info")

	isReplay = true
	logger.Info("replay2 info")

	var messages []string
	for _, log := range observed.AllUntimed() {
		messages = append(messages, log.Message)
	}
	assert.Len(t, messages, 3) // ensures "replay info" wasn't just misspelled
	assert.Contains(t, messages, "normal info")
	assert.NotContains(t, messages, "replay info")
	assert.Contains(t, messages, "normal2 info")
	assert.Contains(t, messages, "replay2 info")

	isReplay = true
	enableLoggingInReplay = true
	parentCore := wrapLogger(&isReplay, &enableLoggingInReplay)
	wrappedCore := parentCore(core).With([]zapcore.Field{zap.String("key", "value")}).(*replayAwareZapCore)
	assert.Equal(t, wrappedCore.isReplay, &isReplay)
	assert.Equal(t, wrappedCore.enableLoggingInReplay, &enableLoggingInReplay)
}

func testDecodeValueHelper(t *testing.T, env *workflowEnvironmentImpl) {
	equals := func(a, b interface{}) bool {
		ao := a.(ActivityOptions)
		bo := b.(ActivityOptions)
		return ao.TaskList == bo.TaskList
	}
	value := ActivityOptions{TaskList: "test-tasklist"}
	blob := env.encodeValue(value)
	isEqual := env.isEqualValue(value, blob, equals)
	require.True(t, isEqual)

	value.TaskList = "value-changed"
	isEqual = env.isEqualValue(value, blob, equals)
	require.False(t, isEqual)
}

func TestDecodedValue(t *testing.T) {
	t.Parallel()
	env := &workflowEnvironmentImpl{
		dataConverter: getDefaultDataConverter(),
	}
	testDecodeValueHelper(t, env)
}

func TestDecodedValue_WithDataConverter(t *testing.T) {
	t.Parallel()
	env := &workflowEnvironmentImpl{
		dataConverter: newTestDataConverter(),
	}
	testDecodeValueHelper(t, env)
}

func Test_DecodedValuePtr(t *testing.T) {
	t.Parallel()
	env := &workflowEnvironmentImpl{
		dataConverter: getDefaultDataConverter(),
	}
	equals := func(a, b interface{}) bool {
		ao := a.(*ActivityOptions)
		bo := b.(*ActivityOptions)
		return ao.TaskList == bo.TaskList
	}
	value := &ActivityOptions{TaskList: "test-tasklist"}
	blob := env.encodeValue(value)
	isEqual := env.isEqualValue(value, blob, equals)
	require.True(t, isEqual)

	value.TaskList = "value-changed"
	isEqual = env.isEqualValue(value, blob, equals)
	require.False(t, isEqual)
}

func Test_DecodedValueNil(t *testing.T) {
	t.Parallel()
	env := &workflowEnvironmentImpl{
		dataConverter: getDefaultDataConverter(),
	}
	equals := func(a, b interface{}) bool {
		return a == nil && b == nil
	}
	// newValue is nil, old value is nil
	var value interface{}
	blob := env.encodeValue(value)
	isEqual := env.isEqualValue(value, blob, equals)
	require.True(t, isEqual)

	// newValue is nil, oldValue is not nil
	blob = env.encodeValue("any-non-nil-value")
	isEqual = env.isEqualValue(value, blob, equals)
	require.False(t, isEqual)

	// newValue is not nil, oldValue is nil
	blob = env.encodeValue(nil)
	isEqual = env.isEqualValue("non-nil-value", blob, equals)
	require.False(t, isEqual)
}

func Test_ValidateAndSerializeSearchAttributes(t *testing.T) {
	t.Parallel()
	_, err := validateAndSerializeSearchAttributes(nil)
	require.EqualError(t, err, "search attributes is empty")

	attr := map[string]interface{}{
		"JustKey": make(chan int),
	}
	_, err = validateAndSerializeSearchAttributes(attr)
	require.EqualError(t, err, "encode search attribute [JustKey] error: json: unsupported type: chan int")

	attr = map[string]interface{}{
		"key": 1,
	}
	searchAttr, err := validateAndSerializeSearchAttributes(attr)
	require.NoError(t, err)
	require.Equal(t, 1, len(searchAttr.IndexedFields))
	var resp int
	json.Unmarshal(searchAttr.IndexedFields["key"], &resp)
	require.Equal(t, 1, resp)
}

func Test_UpsertSearchAttributes(t *testing.T) {
	t.Parallel()
	env := &workflowEnvironmentImpl{
		decisionsHelper: newDecisionsHelper(),
		workflowInfo:    GetWorkflowInfo(createRootTestContext(t)),
	}
	err := env.UpsertSearchAttributes(nil)
	require.Error(t, err)

	err = env.UpsertSearchAttributes(map[string]interface{}{
		CadenceChangeVersion: []string{"change2-1", "change1-1"}},
	)
	require.NoError(t, err)
	_, ok := env.decisionsHelper.decisions[makeDecisionID(decisionTypeUpsertSearchAttributes, "change2-1")]
	require.True(t, ok)
	require.Equal(t, int32(0), env.counterID)

	err = env.UpsertSearchAttributes(map[string]interface{}{"key": 1})
	require.NoError(t, err)
	require.Equal(t, int32(1), env.counterID)
}

func Test_MergeSearchAttributes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		current  *s.SearchAttributes
		upsert   *s.SearchAttributes
		expected *s.SearchAttributes
	}{
		{
			name:     "currentIsNil",
			current:  nil,
			upsert:   &s.SearchAttributes{},
			expected: nil,
		},
		{
			name:     "currentIsEmpty",
			current:  &s.SearchAttributes{IndexedFields: make(map[string][]byte)},
			upsert:   &s.SearchAttributes{},
			expected: nil,
		},
		{
			name: "normalMerge",
			current: &s.SearchAttributes{
				IndexedFields: map[string][]byte{
					"CustomIntField":     []byte(`1`),
					"CustomKeywordField": []byte(`keyword`),
				},
			},
			upsert: &s.SearchAttributes{
				IndexedFields: map[string][]byte{
					"CustomIntField":  []byte(`2`),
					"CustomBoolField": []byte(`true`),
				},
			},
			expected: &s.SearchAttributes{
				IndexedFields: map[string][]byte{
					"CustomIntField":     []byte(`2`),
					"CustomKeywordField": []byte(`keyword`),
					"CustomBoolField":    []byte(`true`),
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := mergeSearchAttributes(test.current, test.upsert)
			require.Equal(t, test.expected, result)
		})
	}
}

func Test_GetChangeVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		changeID string
		version  Version
		expected string
	}{
		{
			name:     "default",
			changeID: "cid",
			version:  DefaultVersion,
			expected: "cid--1",
		},
		{
			name:     "normal_case",
			changeID: "cid",
			version:  1,
			expected: "cid-1",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := getChangeVersion(test.changeID, test.version)
			require.Equal(t, test.expected, result)
		})
	}
}

func Test_GetChangeVersions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		changeID               string
		version                Version
		existingChangeVersions map[string]Version
		expected               []string
	}{
		{
			name:                   "single_change_id",
			changeID:               "cid",
			version:                1,
			existingChangeVersions: map[string]Version{},
			expected:               []string{"cid-1"},
		},
		{
			name:     "multi_change_ids",
			changeID: "cid2",
			version:  1,
			existingChangeVersions: map[string]Version{
				"cid": 1,
			},
			expected: []string{"cid2-1", "cid-1"},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := getChangeVersions(test.changeID, test.version, test.existingChangeVersions)
			require.Equal(t, test.expected, result)
		})
	}
}

func Test_CreateSearchAttributesForChangeVersion(t *testing.T) {
	t.Parallel()
	result := createSearchAttributesForChangeVersion("cid", 1, map[string]Version{})
	val, ok := result["CadenceChangeVersion"]
	require.True(t, ok, "Remember to update related key on server side")
	require.Equal(t, []string{"cid-1"}, val)
}

func TestHistoryEstimationforSmallEvents(t *testing.T) {
	taskList := "tasklist"
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(3),
		{
			EventId:   common.Int64Ptr(4),
			EventType: common.EventTypePtr(s.EventTypeDecisionTaskFailed),
		},
		{
			EventId:   common.Int64Ptr(5),
			EventType: common.EventTypePtr(s.EventTypeWorkflowExecutionSignaled),
		},
		createTestEventDecisionTaskScheduled(6, &s.DecisionTaskScheduledEventAttributes{TaskList: &s.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(7),
	}
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core, zap.Development())
	w := workflowExecutionEventHandlerImpl{
		workflowEnvironmentImpl: &workflowEnvironmentImpl{logger: logger},
	}

	w.logger = logger
	historySizeSum := 0
	for _, event := range testEvents {
		sum := estimateHistorySize(logger, event)
		historySizeSum += sum
	}
	trueSize := len(testEvents) * historySizeEstimationBuffer

	assert.Equal(t, trueSize, historySizeSum)
}

func TestHistoryEstimationforPackedEvents(t *testing.T) {
	// create an array of bytes for testing
	var byteArray []byte
	byteArray = append(byteArray, 100)
	taskList := "tasklist"
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{
			TaskList:                &s.TaskList{Name: &taskList},
			Input:                   byteArray,
			ContinuedFailureDetails: byteArray}),
		createTestEventWorkflowExecutionStarted(2, &s.WorkflowExecutionStartedEventAttributes{
			TaskList:                &s.TaskList{Name: &taskList},
			Input:                   byteArray,
			ContinuedFailureDetails: byteArray}),
		createTestEventWorkflowExecutionStarted(3, &s.WorkflowExecutionStartedEventAttributes{
			TaskList:                &s.TaskList{Name: &taskList},
			Input:                   byteArray,
			ContinuedFailureDetails: byteArray}),
	}
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core, zap.Development())
	w := workflowExecutionEventHandlerImpl{
		workflowEnvironmentImpl: &workflowEnvironmentImpl{logger: logger},
	}

	w.logger = logger
	historySizeSum := 0
	for _, event := range testEvents {
		sum := estimateHistorySize(logger, event)
		historySizeSum += sum
	}
	trueSize := len(testEvents)*historySizeEstimationBuffer + len(byteArray)*2*len(testEvents)
	assert.Equal(t, trueSize, historySizeSum)
}

func TestProcessQuery_KnownQueryTypes(t *testing.T) {
	rootCtx := setWorkflowEnvOptionsIfNotExist(Background())
	eo := getWorkflowEnvOptions(rootCtx)
	eo.queryHandlers["a"] = nil

	weh := &workflowExecutionEventHandlerImpl{
		workflowEnvironmentImpl: &workflowEnvironmentImpl{
			dataConverter: DefaultDataConverter,
		},
		workflowDefinition: &syncWorkflowDefinition{
			rootCtx: rootCtx,
		},
	}

	result, err := weh.ProcessQuery(QueryTypeQueryTypes, nil)
	assert.NoError(t, err)
	assert.Equal(t, "[\"__open_sessions\",\"__query_types\",\"__stack_trace\",\"a\"]\n", string(result))
}

func TestWorkflowExecutionEventHandler_ProcessEvent_WorkflowExecutionStarted(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testRegistry := newRegistry()
		testRegistry.RegisterWorkflowWithOptions(func(ctx Context) error { return nil }, RegisterWorkflowOptions{Name: "test"})

		weh := testWorkflowExecutionEventHandler(t, testRegistry)

		testHeaderStruct := &s.Header{
			Fields: map[string][]byte{
				"test": []byte("test"),
			},
		}
		testInput := []byte("testInput")
		event := &s.HistoryEvent{
			EventType: common.EventTypePtr(s.EventTypeWorkflowExecutionStarted),
			WorkflowExecutionStartedEventAttributes: &s.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &s.WorkflowType{
					Name: common.StringPtr("test"),
				},
				Input:  testInput,
				Header: testHeaderStruct,
			},
		}

		err := weh.ProcessEvent(event, false, false)
		assert.NoError(t, err)
	})
	t.Run("error", func(t *testing.T) {
		testRegistry := newRegistry()

		weh := testWorkflowExecutionEventHandler(t, testRegistry)

		event := &s.HistoryEvent{
			EventType: common.EventTypePtr(s.EventTypeWorkflowExecutionStarted),
			WorkflowExecutionStartedEventAttributes: &s.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &s.WorkflowType{
					Name: common.StringPtr("test"),
				},
			},
		}

		err := weh.ProcessEvent(event, false, false)
		assert.ErrorContains(t, err, errMsgUnknownWorkflowType)
	})
}

func TestWorkflowExecutionEventHandler_ProcessEvent_Noops(t *testing.T) {
	for _, tc := range []s.EventType{
		s.EventTypeWorkflowExecutionCompleted,
		s.EventTypeWorkflowExecutionTimedOut,
		s.EventTypeWorkflowExecutionFailed,
		s.EventTypeDecisionTaskScheduled,
		s.EventTypeDecisionTaskTimedOut,
		s.EventTypeDecisionTaskFailed,
		s.EventTypeActivityTaskStarted,
		s.EventTypeDecisionTaskCompleted,
		s.EventTypeWorkflowExecutionCanceled,
		s.EventTypeWorkflowExecutionContinuedAsNew,
	} {
		t.Run(tc.String(), func(t *testing.T) {
			weh := testWorkflowExecutionEventHandler(t, newRegistry())

			event := &s.HistoryEvent{
				EventType: common.EventTypePtr(tc),
			}

			err := weh.ProcessEvent(event, false, false)
			assert.NoError(t, err)
		})
	}
}

func TestWorkflowExecutionEventHandler_ProcessEvent_nil(t *testing.T) {
	weh := testWorkflowExecutionEventHandler(t, newRegistry())

	err := weh.ProcessEvent(nil, false, false)
	assert.ErrorContains(t, err, "nil event provided")
}

func TestWorkflowExecutionEventHandler_ProcessEvent_no_error_events(t *testing.T) {
	for _, tc := range []struct {
		event          *s.HistoryEvent
		prepareHandler func(*testing.T, *workflowExecutionEventHandlerImpl)
	}{
		{
			event: &s.HistoryEvent{
				EventType: s.EventType(-1).Ptr(),
			},
		},
		{
			event: &s.HistoryEvent{
				EventType: s.EventTypeActivityTaskCompleted.Ptr(),
				EventId:   common.Int64Ptr(2),
				ActivityTaskCompletedEventAttributes: &s.ActivityTaskCompletedEventAttributes{
					Result:           []byte("test"),
					ScheduledEventId: common.Int64Ptr(1),
				},
			},
			prepareHandler: func(t *testing.T, impl *workflowExecutionEventHandlerImpl) {
				decision := impl.decisionsHelper.scheduleActivityTask(&s.ScheduleActivityTaskDecisionAttributes{
					ActivityId: common.StringPtr("test-activity"),
				})
				decision.setData(&scheduledActivity{
					handled: true,
				})
				impl.decisionsHelper.getDecisions(true)
				impl.decisionsHelper.handleActivityTaskScheduled(1, "test-activity")
			},
		},
		{
			event: &s.HistoryEvent{
				EventType: s.EventTypeActivityTaskFailed.Ptr(),
				EventId:   common.Int64Ptr(3),
				ActivityTaskFailedEventAttributes: &s.ActivityTaskFailedEventAttributes{
					Reason:           common.StringPtr("test-reason"),
					ScheduledEventId: common.Int64Ptr(2),
				},
			},
			prepareHandler: func(t *testing.T, impl *workflowExecutionEventHandlerImpl) {
				decision := impl.decisionsHelper.scheduleActivityTask(&s.ScheduleActivityTaskDecisionAttributes{
					ActivityId: common.StringPtr("test-activity"),
				})
				decision.setData(&scheduledActivity{
					handled: false,
					callback: func(result []byte, err error) {
						var customErr *CustomError
						assert.ErrorAs(t, err, &customErr)
						assert.Equal(t, customErr.reason, "test-reason")
					},
				})
				impl.decisionsHelper.getDecisions(true)
				impl.decisionsHelper.handleActivityTaskScheduled(2, "test-activity")
			},
		},
		{
			event: &s.HistoryEvent{
				EventType: s.EventTypeActivityTaskCanceled.Ptr(),
				EventId:   common.Int64Ptr(4),
				ActivityTaskCanceledEventAttributes: &s.ActivityTaskCanceledEventAttributes{
					Details:          []byte("test-details"),
					ScheduledEventId: common.Int64Ptr(3),
				},
			},
			prepareHandler: func(t *testing.T, impl *workflowExecutionEventHandlerImpl) {
				decision := impl.decisionsHelper.scheduleActivityTask(&s.ScheduleActivityTaskDecisionAttributes{
					ActivityId: common.StringPtr("test-activity"),
				})
				decision.setData(&scheduledActivity{
					handled: false,
					callback: func(result []byte, err error) {
						var canceledErr *CanceledError
						assert.ErrorAs(t, err, &canceledErr)
					},
				})
				impl.decisionsHelper.getDecisions(true)
				impl.decisionsHelper.handleActivityTaskScheduled(3, "test-activity")
				decision.cancel()
				decision.handleDecisionSent()
			},
		},
		{
			event: &s.HistoryEvent{
				EventType: s.EventTypeTimerCanceled.Ptr(),
				EventId:   common.Int64Ptr(5),
				TimerCanceledEventAttributes: &s.TimerCanceledEventAttributes{
					TimerId:        common.StringPtr("test-timer"),
					StartedEventId: common.Int64Ptr(4),
				},
			},
			prepareHandler: func(t *testing.T, impl *workflowExecutionEventHandlerImpl) {
				decision := impl.decisionsHelper.startTimer(&s.StartTimerDecisionAttributes{
					TimerId: common.StringPtr("test-timer"),
				})
				decision.setData(&scheduledTimer{
					callback: func(result []byte, err error) {
						var canceledErr *CanceledError
						assert.ErrorAs(t, err, &canceledErr)
					},
				})
				impl.decisionsHelper.getDecisions(true)
				impl.decisionsHelper.handleTimerStarted("test-timer")
				decision.cancel()
				decision.handleDecisionSent()
			},
		},
		{
			event: &s.HistoryEvent{
				EventType: s.EventTypeCancelTimerFailed.Ptr(),
				EventId:   common.Int64Ptr(6),
				CancelTimerFailedEventAttributes: &s.CancelTimerFailedEventAttributes{
					TimerId: common.StringPtr("test-timer"),
					Cause:   common.StringPtr("test-cause"),
				},
			},
			prepareHandler: func(t *testing.T, impl *workflowExecutionEventHandlerImpl) {
				decision := impl.decisionsHelper.startTimer(&s.StartTimerDecisionAttributes{
					TimerId: common.StringPtr("test-timer"),
				})
				decision.setData(&scheduledTimer{
					callback: func(result []byte, err error) {
						var customErr *CustomError
						assert.ErrorAs(t, err, &customErr)
						assert.Equal(t, customErr.reason, "test-cause")
					},
				})
				impl.decisionsHelper.getDecisions(true)
				impl.decisionsHelper.handleTimerStarted("test-timer")
				decision.cancel()
				decision.handleDecisionSent()
			},
		},
		{
			event: &s.HistoryEvent{
				EventType: s.EventTypeWorkflowExecutionCancelRequested.Ptr(),
				EventId:   common.Int64Ptr(7),
				WorkflowExecutionCancelRequestedEventAttributes: &s.WorkflowExecutionCancelRequestedEventAttributes{
					Cause: common.StringPtr("test-cause"),
				},
			},
			prepareHandler: func(t *testing.T, impl *workflowExecutionEventHandlerImpl) {
				cancelCalled := false
				t.Cleanup(func() {
					if !cancelCalled {
						t.Error("cancelWorkflow not called")
						t.FailNow()
					}
				})
				impl.cancelHandler = func() {
					cancelCalled = true
				}
			},
		},
		{
			event: &s.HistoryEvent{
				EventType: s.EventTypeRequestCancelExternalWorkflowExecutionInitiated.Ptr(),
				EventId:   common.Int64Ptr(8),
				RequestCancelExternalWorkflowExecutionInitiatedEventAttributes: &s.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{
					DecisionTaskCompletedEventId: common.Int64Ptr(7),
					Domain:                       common.StringPtr(testDomain),
					WorkflowExecution: &s.WorkflowExecution{
						WorkflowId: common.StringPtr("wid"),
					},
					ChildWorkflowOnly: common.BoolPtr(true),
				},
			},
			prepareHandler: func(t *testing.T, impl *workflowExecutionEventHandlerImpl) {
				decision := impl.decisionsHelper.startChildWorkflowExecution(&s.StartChildWorkflowExecutionDecisionAttributes{
					Domain:     common.StringPtr(testDomain),
					WorkflowId: common.StringPtr("wid"),
				})
				decision.setData(&scheduledChildWorkflow{
					handled: true,
				})
				impl.decisionsHelper.getDecisions(true)
				impl.decisionsHelper.handleStartChildWorkflowExecutionInitiated("wid")
				impl.decisionsHelper.handleChildWorkflowExecutionStarted("wid")
				decision = impl.decisionsHelper.requestCancelExternalWorkflowExecution(testDomain, "wid", "", "", true)
				decision.setData(&scheduledCancellation{
					callback: func(result []byte, err error) {
						var canceledErr *CanceledError
						assert.ErrorAs(t, err, &canceledErr)
					},
				})
				impl.decisionsHelper.getDecisions(true)
				impl.decisionsHelper.handleRequestCancelExternalWorkflowExecutionInitiated(7, "wid", "")
				decision.handleDecisionSent()
			},
		},
	} {
		t.Run(tc.event.EventType.String(), func(t *testing.T) {
			weh := testWorkflowExecutionEventHandler(t, newRegistry())
			if tc.prepareHandler != nil {
				tc.prepareHandler(t, weh)
			}

			// EventHandle handles all internal failures as panics.
			// make sure we don't fail event processing.
			// This can happen if there is a panic inside the handler.
			weh.completeHandler = func(result []byte, err error) {
				assert.NoError(t, err)
			}

			err := weh.ProcessEvent(tc.event, false, false)
			assert.NoError(t, err)
		})
	}
}

func TestSideEffect(t *testing.T) {
	t.Run("replay", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		weh.workflowEnvironmentImpl.isReplay = true
		weh.sideEffectResult[weh.counterID] = []byte("test")
		weh.SideEffect(func() ([]byte, error) {
			t.Error("side effect should not be called during replay")
			t.Failed()
			return nil, assert.AnError
		}, func(result []byte, err error) {
			assert.NoError(t, err)
			assert.Equal(t, "test", string(result))
		})
	})
	t.Run("success", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		weh.SideEffect(func() ([]byte, error) {
			return []byte("test"), nil
		}, func(result []byte, err error) {
			assert.NoError(t, err)
			assert.Equal(t, "test", string(result))
		})
	})
}

func TestGetVersion_validation(t *testing.T) {
	t.Run("version < minSupported", func(t *testing.T) {
		assert.PanicsWithValue(t, `Workflow code removed support of version 1. for "test" changeID. The oldest supported version is 2`, func() {
			validateVersion("test", 1, 2, 3)
		})
	})
	t.Run("version > maxSupported", func(t *testing.T) {
		assert.PanicsWithValue(t, `Workflow code is too old to support version 3 for "test" changeID. The maximum supported version is 2`, func() {
			validateVersion("test", 3, 1, 2)
		})
	})
	t.Run("success", func(t *testing.T) {
		validateVersion("test", 2, 1, 3)
	})
}

func TestGetVersion(t *testing.T) {
	t.Run("version exists", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		weh.changeVersions = map[string]Version{
			"test": 2,
		}
		res := weh.GetVersion("test", 1, 3)
		assert.Equal(t, Version(2), res)
	})
	t.Run("version exists, ExecuteWithVersion is used", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		weh.changeVersions = map[string]Version{
			"test": 2,
		}
		res := weh.GetVersion("test", 1, 3, ExecuteWithVersion(3))
		assert.Equal(t, Version(2), res)
	})
	t.Run("version exists, ExecuteWithMinVersion is used", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		weh.changeVersions = map[string]Version{
			"test": 2,
		}
		res := weh.GetVersion("test", 1, 3, ExecuteWithMinVersion())
		assert.Equal(t, Version(2), res)
	})
	t.Run("version doesn't exist in replay", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		weh.isReplay = true
		res := weh.GetVersion("test", DefaultVersion, 3)
		assert.Equal(t, DefaultVersion, res)
		require.Contains(t, weh.changeVersions, "test")
		assert.Equal(t, DefaultVersion, weh.changeVersions["test"])
	})
	t.Run("version doesn't exist without replay", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		res := weh.GetVersion("test", DefaultVersion, 3)
		assert.Equal(t, Version(3), res)
		require.Contains(t, weh.changeVersions, "test")
		assert.Equal(t, Version(3), weh.changeVersions["test"])
		assert.Equal(t, []byte(`["test-3"]`), weh.workflowInfo.SearchAttributes.IndexedFields[CadenceChangeVersion], "ensure search attributes are updated")
	})
	t.Run("version doesn't exist, ExecuteWithVersion is used", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		res := weh.GetVersion("test", DefaultVersion, 3, ExecuteWithVersion(2))
		assert.Equal(t, Version(2), res)
		require.Contains(t, weh.changeVersions, "test")
		assert.Equal(t, Version(2), weh.changeVersions["test"])
		assert.Equal(t, []byte(`["test-2"]`), weh.workflowInfo.SearchAttributes.IndexedFields[CadenceChangeVersion], "ensure search attributes are updated")
	})
	t.Run("version doesn't exist, ExecuteWithVersion is used, DefaultVersion is used", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		res := weh.GetVersion("test", DefaultVersion, 3, ExecuteWithVersion(DefaultVersion))
		assert.Equal(t, DefaultVersion, res)
		require.Contains(t, weh.changeVersions, "test")
		assert.Equal(t, DefaultVersion, weh.changeVersions["test"])
		require.Nil(t, weh.workflowInfo.SearchAttributes, "ensure search attributes are not updated")
	})
	t.Run("version doesn't exist, ExecuteWithMinVersion is used, min is non DefaultVersion", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		res := weh.GetVersion("test", 1, 3, ExecuteWithMinVersion())
		assert.Equal(t, Version(1), res)
		require.Contains(t, weh.changeVersions, "test")
		assert.Equal(t, Version(1), weh.changeVersions["test"])
		assert.Equal(t, []byte(`["test-1"]`), weh.workflowInfo.SearchAttributes.IndexedFields[CadenceChangeVersion], "ensure search attributes are updated")
	})
	t.Run("version doesn't exist, ExecuteWithMinVersion is used, DefaultVersion is used", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		res := weh.GetVersion("test", DefaultVersion, 3, ExecuteWithMinVersion())
		assert.Equal(t, DefaultVersion, res)
		require.Contains(t, weh.changeVersions, "test")
		assert.Equal(t, DefaultVersion, weh.changeVersions["test"])
		require.Nil(t, weh.workflowInfo.SearchAttributes, "ensure search attributes are not updated")
	})

	t.Run("version doesn't exist, ExecuteWithVersion is used, version > maximum version", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		assert.PanicsWithValue(t, `Workflow code is too old to support version 10 for "test" changeID. The maximum supported version is 3`, func() {
			weh.GetVersion("test", DefaultVersion, 3, ExecuteWithVersion(10))
		})

		require.Nil(t, weh.workflowInfo.SearchAttributes, "ensure search attributes are not updated")
	})
	t.Run("version doesn't exist, ExecuteWithVersion is used, version < minimum version", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		assert.PanicsWithValue(t, `Workflow code removed support of version 0. for "test" changeID. The oldest supported version is 1`, func() {
			weh.GetVersion("test", 1, 3, ExecuteWithVersion(0))
		})

		require.Nil(t, weh.workflowInfo.SearchAttributes, "ensure search attributes are not updated")
	})
}

func TestMutableSideEffect(t *testing.T) {
	t.Run("replay with existing value", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		weh.mutableSideEffect["test-id"] = []byte(`"existing-value"`)
		weh.isReplay = true

		result := weh.MutableSideEffect("test-id", func() interface{} {
			t.Error("side effect function should not be called during replay with existing value")
			return "new-value"
		}, func(a, b interface{}) bool {
			return a == b
		})
		var value string
		err := result.Get(&value)
		assert.NoError(t, err)
		assert.Equal(t, "existing-value", value)
	})

	t.Run("replay without existing value", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		weh.isReplay = true
		assert.PanicsWithValue(t, "Non deterministic workflow code change detected. MutableSideEffect API call doesn't have a correspondent event in the workflow history. MutableSideEffect ID: test-id", func() {
			weh.MutableSideEffect("test-id", func() interface{} {
				return "new-value"
			}, func(a, b interface{}) bool {
				return a == b
			})
		})
	})
	t.Run("non-replay without value", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())

		result := weh.MutableSideEffect("test-id", func() interface{} {
			return "existing-value"
		}, func(a, b interface{}) bool {
			return a == b
		})

		var value string
		err := result.Get(&value)
		assert.NoError(t, err)
		assert.Equal(t, "existing-value", value)
	})
	t.Run("non-replay with equal value", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		weh.mutableSideEffect["test-id"] = []byte(`"existing-value"`)

		result := weh.MutableSideEffect("test-id", func() interface{} {
			return "existing-value"
		}, func(a, b interface{}) bool {
			return a == b
		})

		var value string
		err := result.Get(&value)
		assert.NoError(t, err)
		assert.Equal(t, "existing-value", value)
	})
	t.Run("non-replay with different value", func(t *testing.T) {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		weh.mutableSideEffect["test-id"] = []byte(`"existing-value"`)

		result := weh.MutableSideEffect("test-id", func() interface{} {
			return "new-value"
		}, func(a, b interface{}) bool {
			return a == b
		})

		var value string
		err := result.Get(&value)
		assert.NoError(t, err)
		assert.Equal(t, "new-value", value)
		// the last symbol is a control symbol, so we need to check the value without it.
		assert.Equal(t, []byte(`"new-value"`), weh.mutableSideEffect["test-id"][:len(weh.mutableSideEffect["test-id"])-1])
	})
}

func TestEventHandler_handleMarkerRecorded(t *testing.T) {
	for _, tc := range []struct {
		marker       *s.MarkerRecordedEventAttributes
		assertResult func(t *testing.T, result *workflowExecutionEventHandlerImpl)
	}{
		{
			marker: &s.MarkerRecordedEventAttributes{
				MarkerName: common.StringPtr(sideEffectMarkerName),
				Details:    getSerializedDetails(t, 1, []byte("test")),
			},
			assertResult: func(t *testing.T, result *workflowExecutionEventHandlerImpl) {
				require.Contains(t, result.sideEffectResult, int32(1))
				assert.Equal(t, []byte("test"), result.sideEffectResult[1])
			},
		},
		{
			marker: &s.MarkerRecordedEventAttributes{
				MarkerName: common.StringPtr(versionMarkerName),
				Details:    getSerializedDetails(t, "test-version", Version(1)),
			},
			assertResult: func(t *testing.T, result *workflowExecutionEventHandlerImpl) {
				require.Contains(t, result.changeVersions, "test-version")
				assert.Equal(t, Version(1), result.changeVersions["test-version"])
			},
		},
		{
			marker: &s.MarkerRecordedEventAttributes{
				MarkerName: common.StringPtr(mutableSideEffectMarkerName),
				Details:    getSerializedDetails(t, "test-marker", "test"),
			},
			assertResult: func(t *testing.T, result *workflowExecutionEventHandlerImpl) {
				require.Contains(t, result.mutableSideEffect, "test-marker")
				assert.Equal(t, []byte("test"), result.mutableSideEffect["test-marker"])
			},
		},
	} {
		weh := testWorkflowExecutionEventHandler(t, newRegistry())
		err := weh.handleMarkerRecorded(1, tc.marker)
		assert.NoError(t, err)
	}
}

func TestEventHandler_handleMarkerRecorded_failures(t *testing.T) {
	for _, tc := range []struct {
		name           string
		marker         *s.MarkerRecordedEventAttributes
		assertErrorStr string
	}{
		{
			name: "unknown marker",
			marker: &s.MarkerRecordedEventAttributes{
				MarkerName: common.StringPtr("unknown"),
			},
			assertErrorStr: "unknown marker name \"unknown\" for eventID \"1\"",
		},
		{
			name: "side effect with invalid details",
			marker: &s.MarkerRecordedEventAttributes{
				MarkerName: common.StringPtr(sideEffectMarkerName),
				Details:    []byte("invalid"),
			},
			assertErrorStr: "extract side effect: unable to decode argument:",
		},
		{
			name: "version with invalid details",
			marker: &s.MarkerRecordedEventAttributes{
				MarkerName: common.StringPtr(versionMarkerName),
				Details:    []byte("invalid"),
			},
			assertErrorStr: "extract change id: unable to decode argument:",
		},
		{
			name: "mutable side effect with invalid details",
			marker: &s.MarkerRecordedEventAttributes{
				MarkerName: common.StringPtr(mutableSideEffectMarkerName),
				Details:    []byte("invalid"),
			},
			assertErrorStr: "extract fixed id: unable to decode argument:",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			weh := testWorkflowExecutionEventHandler(t, newRegistry())
			err := weh.handleMarkerRecorded(1, tc.marker)
			assert.ErrorContains(t, err, tc.assertErrorStr)
		})
	}
}

func TestWorkflowEnvironment_sessions(t *testing.T) {
	handler := testWorkflowExecutionEventHandler(t, newRegistry())
	testSession := &SessionInfo{
		SessionID: "test-session",
		HostName:  "test-host",
	}
	handler.AddSession(testSession)
	list := handler.getOpenSessions()
	assert.Contains(t, list, testSession)
	handler.RemoveSession(testSession.SessionID)
	list = handler.getOpenSessions()
	assert.Empty(t, list)
}

func TestWorkflowExecutionEnvironment_NewTimer_immediate_calls(t *testing.T) {
	t.Run("immediate call", func(t *testing.T) {
		handler := testWorkflowExecutionEventHandler(t, newRegistry())
		handlerCalled := false
		res := handler.NewTimer(0, func(result []byte, err error) {
			assert.NoError(t, err)
			handlerCalled = true
		})
		assert.True(t, handlerCalled, "handler must be called immediately")
		assert.Nil(t, res)
	})
	t.Run("negative duration", func(t *testing.T) {
		handler := testWorkflowExecutionEventHandler(t, newRegistry())
		handlerCalled := false
		res := handler.NewTimer(-2*time.Second, func(result []byte, err error) {
			handlerCalled = true
			assert.ErrorContains(t, err, "negative duration provided")
		})
		assert.Nil(t, res)
		assert.True(t, handlerCalled, "handler must be called immediately")
	})
	t.Run("timer cancellation", func(t *testing.T) {
		handler := testWorkflowExecutionEventHandler(t, newRegistry())
		timer := handler.NewTimer(time.Second, func(result []byte, err error) {
			assert.ErrorIs(t, err, ErrCanceled)
		})
		handler.RequestCancelTimer(timer.timerID)
	})
}

func testWorkflowExecutionEventHandler(t *testing.T, registry *registry) *workflowExecutionEventHandlerImpl {
	var testWorkflowInfo = &WorkflowInfo{
		WorkflowType: WorkflowType{
			Name: "test",
			Path: "",
		},
	}

	return newWorkflowExecutionEventHandler(
		testWorkflowInfo,
		func(result []byte, err error) {},
		testlogger.NewZap(t),
		true,
		tally.NewTestScope("test", nil),
		registry,
		&defaultDataConverter{},
		nil,
		opentracing.NoopTracer{},
		nil,
	).(*workflowExecutionEventHandlerImpl)
}

func getSerializedDetails[T, V any](t *testing.T, id T, data V) []byte {
	converter := defaultDataConverter{}
	res, err := converter.ToData(id, data)
	require.NoError(t, err)
	return res
}
