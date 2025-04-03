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
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
	"github.com/uber/tchannel-go/thrift"
	"go.uber.org/yarpc"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	s "go.uber.org/cadence/.gen/go/shared"
)

var (
	safeCharacters = []rune{'_'}

	sanitizeOptions = tally.SanitizeOptions{
		NameCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		KeyCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		ValueCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		ReplacementCharacter: tally.DefaultReplacementCharacter,
	}
)

type testCase struct {
	serviceMethod string
	callArgs      []interface{}
}

type errCase struct {
	err              error
	expectedCounters []string
}

func Test_Wrapper(t *testing.T) {
	ctx, _ := thrift.NewContext(time.Minute)

	typeWrapper := reflect.TypeOf(&workflowServiceMetricsWrapper{})
	tests := make([]testCase, 0, typeWrapper.NumMethod())
	for numMethod := 0; numMethod < typeWrapper.NumMethod(); numMethod++ {
		method := typeWrapper.Method(numMethod)
		inputs := make([]interface{}, 0, method.Type.NumIn()-1)
		inputs = append(inputs, ctx)

		for i := 1; i < method.Type.NumIn(); i++ {
			if method.Type.In(i).Kind() == reflect.Ptr {
				inputs = append(inputs, reflect.New(method.Type.In(i).Elem()).Interface())
			}
		}

		tests = append(tests, testCase{
			serviceMethod: method.Name,
			callArgs:      inputs,
		})
	}

	// run each test twice - once with the regular scope, once with a sanitized metrics scope
	for _, test := range tests {
		for _, errCase := range []struct {
			err              error
			expectedCounters []string
		}{
			{
				err:              nil,
				expectedCounters: []string{CadenceRequest},
			},
			{
				err:              &s.EntityNotExistsError{},
				expectedCounters: []string{CadenceRequest, CadenceInvalidRequest},
			},
			{
				err:              &s.InternalServiceError{},
				expectedCounters: []string{CadenceRequest, CadenceError},
			},
		} {
			runTest(t, test, errCase, newService, assertMetrics, fmt.Sprintf("%v_errcase_%v_normal", test.serviceMethod, errCase.err))
			runTest(t, test, errCase, newPromService, assertPromMetrics, fmt.Sprintf("%v_errcase_%v_prom_sanitized", test.serviceMethod, errCase.err))
		}

	}
}

func runTest(
	t *testing.T,
	test testCase,
	errCase errCase,
	serviceFunc func(*testing.T) (*workflowservicetest.MockClient, workflowserviceclient.Interface, io.Closer, *CapturingStatsReporter),
	validationFunc func(*testing.T, *CapturingStatsReporter, string, []string),
	name string,
) {
	t.Run(name, func(t *testing.T) {
		mockService, wrapperService, closer, reporter := serviceFunc(t)
		mockServiceVal := reflect.ValueOf(mockService)
		method, exists := mockServiceVal.Type().MethodByName(test.serviceMethod)
		require.True(t, exists, "method %s does not exists", test.serviceMethod)

		expecterVals := mockServiceVal.MethodByName("EXPECT").Call(nil)
		expectedMethod := expecterVals[0].MethodByName(test.serviceMethod)
		mockInputs := make([]reflect.Value, 0, expectedMethod.Type().NumIn())
		for inCounter := 0; inCounter < expectedMethod.Type().NumIn(); inCounter++ {
			mockInputs = append(mockInputs, reflect.ValueOf(gomock.Any()))
		}
		callVals := expectedMethod.Call(mockInputs)

		returnVals := make([]reflect.Value, 0, method.Type.NumOut())
		for i := 0; i < method.Type.NumOut()-1; i++ {
			output := method.Type.Out(i)
			if errCase.err == nil {
				returnVals = append(returnVals, reflect.New(output).Elem())
			} else {
				returnVals = append(returnVals, reflect.Zero(output))
			}
		}

		if errCase.err != nil {
			returnVals = append(returnVals, reflect.ValueOf(errCase.err))
		} else {
			returnVals = append(returnVals, nilError)
		}

		callVals[0].MethodByName("Return").Call(returnVals)

		callOption := yarpc.CallOption{}
		inputs := make([]reflect.Value, len(test.callArgs))
		for i, arg := range test.callArgs {
			inputs[i] = reflect.ValueOf(arg)
		}
		inputs = append(inputs, reflect.ValueOf(callOption))
		actualMethod := reflect.ValueOf(wrapperService).MethodByName(test.serviceMethod)
		methodReturnVals := actualMethod.Call(inputs)
		err := methodReturnVals[len(methodReturnVals)-1].Interface()
		if errCase.err != nil {
			assert.ErrorIs(t, err.(error), errCase.err)
		} else {
			assert.Nil(t, err, "error must be nil")
		}
		require.NoError(t, closer.Close())
		validationFunc(t, reporter, test.serviceMethod, errCase.expectedCounters)
	})
}

func assertMetrics(t *testing.T, reporter *CapturingStatsReporter, methodName string, counterNames []string) {
	assert.Equal(t, len(counterNames), len(reporter.counts), "expected %v counters, got %v", counterNames, reporter.counts)
	for _, name := range counterNames {
		counterName := CadenceMetricsPrefix + methodName + "." + name
		find := false
		// counters are not in order
		for _, counter := range reporter.counts {
			if counterName == counter.name {
				find = true
				break
			}
		}
		assert.True(t, find, "counter %v not found in counters %v", counterName, reporter.counts)
	}
	assert.Equal(t, 1, len(reporter.timers), "expected 1 timer, got %v", len(reporter.timers))
	assert.Equal(t, CadenceMetricsPrefix+methodName+"."+CadenceLatency, reporter.timers[0].name, "expected timer %v, got %v", CadenceMetricsPrefix+methodName+"."+CadenceLatency, reporter.timers[0].name)
}

func assertPromMetrics(t *testing.T, reporter *CapturingStatsReporter, methodName string, counterNames []string) {
	assert.Equal(t, len(counterNames), len(reporter.counts), "expected %v counters, got %v", counterNames, reporter.counts)
	for _, name := range counterNames {
		counterName := makePromCompatible(CadenceMetricsPrefix + methodName + "." + name)
		find := false
		// counters are not in order
		for _, counter := range reporter.counts {
			if counterName == counter.name {
				find = true
				break
			}
		}
		assert.True(t, find, "counter %v not found in counters %v", counterName, reporter.counts)
	}
	assert.Equal(t, 1, len(reporter.timers), "expected 1 timer, got %v", len(reporter.timers))
	expected := makePromCompatible(CadenceMetricsPrefix + methodName + "." + CadenceLatency)
	assert.Equal(t, expected, reporter.timers[0].name, "expected timer %v, got %v", expected, reporter.timers[0].name)
}

func makePromCompatible(name string) string {
	name = strings.Replace(name, "-", "_", -1)
	name = strings.Replace(name, ".", "_", -1)
	return name
}

func newService(t *testing.T) (
	mockService *workflowservicetest.MockClient,
	wrapperService workflowserviceclient.Interface,
	closer io.Closer,
	reporter *CapturingStatsReporter,
) {
	mockCtrl := gomock.NewController(t)
	mockService = workflowservicetest.NewMockClient(mockCtrl)
	isReplay := false
	scope, closer, reporter := NewMetricsScope(&isReplay)
	wrapperService = NewWorkflowServiceWrapper(mockService, scope)
	return
}

func newPromService(t *testing.T) (
	mockService *workflowservicetest.MockClient,
	wrapperService workflowserviceclient.Interface,
	closer io.Closer,
	reporter *CapturingStatsReporter,
) {
	mockCtrl := gomock.NewController(t)
	mockService = workflowservicetest.NewMockClient(mockCtrl)
	isReplay := false
	scope, closer, reporter := newPromScope(&isReplay)
	wrapperService = NewWorkflowServiceWrapper(mockService, scope)
	return
}

func newPromScope(isReplay *bool) (tally.Scope, io.Closer, *CapturingStatsReporter) {
	reporter := &CapturingStatsReporter{}
	opts := tally.ScopeOptions{
		Reporter:        reporter,
		SanitizeOptions: &sanitizeOptions,
	}
	scope, closer := tally.NewRootScope(opts, time.Second)
	return WrapScope(isReplay, scope, &realClock{}), closer, reporter
}

var nilError = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())
