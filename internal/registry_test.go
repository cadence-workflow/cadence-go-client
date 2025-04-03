// Copyright (c) 2017-2020 Uber Technologies Inc.
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

	"github.com/stretchr/testify/require"
)

func TestWorkflowRegistration(t *testing.T) {
	w := &testWorkflowStruct{}
	tests := []struct {
		msg               string
		register          func(r *registry)
		registerPanic     bool
		workflowType      string
		altWorkflowType   string
		resolveByFunction interface{}
		resolveByAlias    string
	}{
		{
			msg:               "register workflow function",
			register:          func(r *registry) { r.RegisterWorkflow(testWorkflowFunction) },
			workflowType:      "go.uber.org/cadence/internal.testWorkflowFunction",
			resolveByFunction: testWorkflowFunction,
		},
		{
			msg: "register workflow function with short name",
			register: func(r *registry) {
				r.RegisterWorkflowWithOptions(testWorkflowFunction, RegisterWorkflowOptions{EnableShortName: true})
			},
			workflowType:      "testWorkflowFunction",
			resolveByFunction: testWorkflowFunction,
		},
		{
			msg: "register workflow function with alias",
			register: func(r *registry) {
				r.RegisterWorkflowWithOptions(testWorkflowFunction, RegisterWorkflowOptions{Name: "workflow.alias"})
			},
			workflowType:      "workflow.alias",
			resolveByFunction: testWorkflowFunction,
			resolveByAlias:    "workflow.alias",
		},
		{
			msg:               "register workflow struct function (backwards compatible)",
			register:          func(r *registry) { r.RegisterWorkflow(w.Method) },
			workflowType:      "go.uber.org/cadence/internal.(*testWorkflowStruct).Method",
			altWorkflowType:   "go.uber.org/cadence/internal.(*testWorkflowStruct).Method-fm",
			resolveByFunction: w.Method,
		},
		{
			msg: "register duplicated workflow in one registry (should panic)",
			register: func(r *registry) {
				r.RegisterWorkflow(testWorkflowFunction)
				r.RegisterWorkflow(testWorkflowFunction)
			},
			registerPanic: true,
		},
		{
			msg: "register duplicated workflow with already registered check disabled",
			register: func(r *registry) {
				r.RegisterWorkflow(testWorkflowFunction)
				r.RegisterWorkflowWithOptions(testWorkflowFunction, RegisterWorkflowOptions{DisableAlreadyRegisteredCheck: true})
			},
			workflowType:      "go.uber.org/cadence/internal.testWorkflowFunction",
			resolveByFunction: testWorkflowFunction,
		},
		{
			msg: "register duplicated workflow in chained registry (should panic)",
			register: func(r *registry) {
				r.next.RegisterWorkflow(testWorkflowFunction)
				r.RegisterWorkflow(testWorkflowFunction)
			},
			registerPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			r := newRegistry()
			if tt.registerPanic {
				require.Panics(t, func() { tt.register(r) }, "register should panic")
				return
			}

			tt.register(r)

			// Verify registered workflow type
			workflowType := r.GetRegisteredWorkflows()[0].WorkflowType().Name
			require.Equal(t, tt.workflowType, workflowType)

			// Verify workflow is resolved from workflow type
			_, ok := r.getWorkflowFn(tt.workflowType)
			require.True(t, ok)

			// Verify workflow is resolved from alternative (backwards compatible) workflow type
			if len(tt.altWorkflowType) > 0 {
				_, ok := r.getWorkflowFn(tt.altWorkflowType)
				require.True(t, ok)
			}

			// Verify resolving by function reference
			workflowType = getWorkflowFunctionName(r, tt.resolveByFunction)
			require.Equal(t, tt.workflowType, workflowType)

			// Verify resolving by alias
			if tt.resolveByAlias != "" {
				workflowType = getWorkflowFunctionName(r, tt.resolveByAlias)
				require.Equal(t, tt.workflowType, workflowType)
			}
		})
	}
}

func TestActivityRegistration(t *testing.T) {
	tests := []struct {
		msg               string
		register          func(r *registry)
		registerPanic     bool
		activityType      string
		altActivityType   string
		resolveByFunction interface{}
		resolveByAlias    string
	}{
		{
			msg:               "register activity function",
			register:          func(r *registry) { r.RegisterActivity(testActivityFunction) },
			activityType:      "go.uber.org/cadence/internal.testActivityFunction",
			resolveByFunction: testActivityFunction,
		},
		{
			msg: "register activity function with short name",
			register: func(r *registry) {
				r.RegisterActivityWithOptions(testActivityFunction, RegisterActivityOptions{EnableShortName: true})
			},
			activityType:      "testActivityFunction",
			resolveByFunction: testActivityFunction,
		},
		{
			msg: "register activity function with an alias",
			register: func(r *registry) {
				r.RegisterActivityWithOptions(testActivityFunction, RegisterActivityOptions{Name: "activity.alias"})
			},
			activityType:      "activity.alias",
			resolveByFunction: testActivityFunction,
			resolveByAlias:    "activity.alias",
		},
		{
			msg:               "register activity struct",
			register:          func(r *registry) { r.RegisterActivity(&testActivityStruct{}) },
			activityType:      "go.uber.org/cadence/internal.(*testActivityStruct).Method",
			altActivityType:   "go.uber.org/cadence/internal.(*testActivityStruct).Method-fm",
			resolveByFunction: (&testActivityStruct{}).Method,
		},
		{
			msg: "register activity struct with short name",
			register: func(r *registry) {
				r.RegisterActivityWithOptions(&testActivityStruct{}, RegisterActivityOptions{EnableShortName: true})
			},
			activityType:      "Method",
			resolveByFunction: (&testActivityStruct{}).Method,
		},
		{
			msg: "register activity struct with a prefix",
			register: func(r *registry) {
				r.RegisterActivityWithOptions(&testActivityStruct{}, RegisterActivityOptions{Name: "prefix."})
			},
			activityType:      "prefix.Method",
			resolveByFunction: (&testActivityStruct{}).Method,
			resolveByAlias:    "prefix.Method",
		},
		{
			msg: "register duplicated activity function in one registry (should panic)",
			register: func(r *registry) {
				duplicatedActivityAlias := "activity.alias"
				r.RegisterActivityWithOptions(testActivityFunction, RegisterActivityOptions{Name: duplicatedActivityAlias})
				r.RegisterActivityWithOptions(testActivityFunction, RegisterActivityOptions{Name: duplicatedActivityAlias})
			},
			registerPanic: true,
		},
		{
			msg: "register duplicated activity struct with already registered check disabled",
			register: func(r *registry) {
				r.RegisterActivity(&testActivityStruct{})
				r.RegisterActivityWithOptions(&testActivityStruct{}, RegisterActivityOptions{DisableAlreadyRegisteredCheck: true})
			},
			activityType:      "go.uber.org/cadence/internal.(*testActivityStruct).Method",
			resolveByFunction: (&testActivityStruct{}).Method,
		},
		{
			msg: "register duplicated activity function in chained registry (should panic)",
			register: func(r *registry) {
				r.next.RegisterActivity(testActivityFunction)
				r.RegisterActivity(testActivityFunction)
			},
			registerPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			r := newRegistry()
			if tt.registerPanic {
				require.Panics(t, func() { tt.register(r) }, "register should panic")
				return
			}

			tt.register(r)

			// Verify registered activity type
			activityType := r.getRegisteredActivities()[0].ActivityType().Name
			require.Equal(t, tt.activityType, activityType, "activity type")

			// Verify activity is resolved from activity type
			_, ok := r.GetActivity(tt.activityType)
			require.True(t, ok)

			// Verify activity is resolved from alternative (backwards compatible) activity type
			if len(tt.altActivityType) > 0 {
				_, ok = r.GetActivity(tt.altActivityType)
				require.True(t, ok)
			}

			// Verify resolving by function reference
			activityType = getActivityFunctionName(r, tt.resolveByFunction)
			require.Equal(t, tt.activityType, activityType, "resolve by function reference")

			// Verify resolving by alias
			if tt.resolveByAlias != "" {
				activityType = getActivityFunctionName(r, tt.resolveByAlias)
				require.Equal(t, tt.activityType, activityType, "resolve by alias")
			}
		})
	}
}

type testWorkflowStruct struct{}
type testActivityStruct struct{}

func (ts *testWorkflowStruct) Method(ctx Context) error { return nil }
func (ts *testActivityStruct) Method() error            { return nil }

func testActivityFunction() error            { return nil }
func testWorkflowFunction(ctx Context) error { return nil }
