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
	"fmt"
	"reflect"
	"strings"
	"sync"
)

const (
	errMsgUnknownWorkflowType = "unable to find workflow type"
)

var once sync.Once

// Singleton to hold the host registration details.
var globalRegistry *registry

func newRegistry() *registry {
	return &registry{
		workflowFuncMap:  make(map[string]workflow),
		workflowAliasMap: make(map[string]string),
		activityFuncMap:  make(map[string]activity),
		activityAliasMap: make(map[string]string),
		next:             getGlobalRegistry(),
	}
}

func getGlobalRegistry() *registry {
	once.Do(func() {
		globalRegistry = &registry{
			workflowFuncMap:  make(map[string]workflow),
			workflowAliasMap: make(map[string]string),
			activityFuncMap:  make(map[string]activity),
			activityAliasMap: make(map[string]string),
		}
	})
	return globalRegistry
}

type registry struct {
	sync.Mutex
	workflowFuncMap  map[string]workflow
	workflowAliasMap map[string]string
	activityFuncMap  map[string]activity
	activityAliasMap map[string]string
	next             *registry // Allows to chain registries
}

func (r *registry) RegisterWorkflow(af interface{}) {
	r.RegisterWorkflowWithOptions(af, RegisterWorkflowOptions{})
}

func (r *registry) RegisterWorkflowWithOptions(
	wf interface{},
	options RegisterWorkflowOptions,
) {
	// Validate that it is a function
	fnType := reflect.TypeOf(wf)
	if err := validateFnFormat(fnType, true); err != nil {
		panic(err)
	}
	fnName := getFunctionName(wf)
	alias := options.Name
	registerName := fnName

	if options.EnableShortName {
		registerName = getShortFunctionName(fnName)
	}
	if len(alias) > 0 {
		registerName = alias
	}

	r.Lock()
	defer r.Unlock()

	if !options.DisableAlreadyRegisteredCheck {
		if _, ok := r.getWorkflowNoLock(registerName); ok {
			panic(fmt.Sprintf("workflow name \"%v\" is already registered", registerName))
		}
	}
	r.workflowFuncMap[registerName] = &workflowExecutor{registerName, wf, fnName}
	if len(alias) > 0 || options.EnableShortName {
		r.workflowAliasMap[fnName] = registerName
	}
}

func (r *registry) RegisterActivity(af interface{}) {
	r.RegisterActivityWithOptions(af, RegisterActivityOptions{})
}

func (r *registry) RegisterActivityWithOptions(af interface{}, options RegisterActivityOptions) {
	fnType := reflect.TypeOf(af)
	var err error
	if fnType.Kind() == reflect.Ptr && fnType.Elem().Kind() == reflect.Struct {
		err = r.registerActivityStruct(af, options)
	} else {
		err = r.registerActivityFunction(af, options)
	}

	if err != nil {
		panic(err)
	}
}

func (r *registry) GetRegisteredWorkflows() []workflow {
	r.Lock()
	var result []workflow
	for _, wf := range r.workflowFuncMap {
		result = append(result, wf)
	}
	r.Unlock()
	if r.next != nil {
		nextWFs := r.next.GetRegisteredWorkflows()
		result = append(result, nextWFs...)
	}
	return result
}

func (r *registry) registerActivityFunction(af interface{}, options RegisterActivityOptions) error {
	fnType := reflect.TypeOf(af)
	if err := validateFnFormat(fnType, false); err != nil {
		return fmt.Errorf("failed to register activity method: %v", err)
	}

	fnName := getFunctionName(af)
	alias := options.Name
	registerName := fnName

	if options.EnableShortName {
		registerName = getShortFunctionName(fnName)
	}
	if len(alias) > 0 {
		registerName = alias
	}

	r.Lock()
	defer r.Unlock()

	if !options.DisableAlreadyRegisteredCheck {
		if _, ok := r.getActivityNoLock(registerName); ok {
			return fmt.Errorf("activity type \"%v\" is already registered", registerName)
		}
	}
	r.activityFuncMap[registerName] = &activityExecutor{registerName, af, options, fnName}
	if len(alias) > 0 || options.EnableShortName {
		r.activityAliasMap[fnName] = registerName
	}

	return nil
}

func (r *registry) registerActivityStruct(aStruct interface{}, options RegisterActivityOptions) error {
	r.Lock()
	defer r.Unlock()

	structValue := reflect.ValueOf(aStruct)
	structType := structValue.Type()
	count := 0
	for i := 0; i < structValue.NumMethod(); i++ {
		methodValue := structValue.Method(i)
		method := structType.Method(i)
		// skip private method
		if method.PkgPath != "" {
			continue
		}
		methodName := getFunctionName(method.Func.Interface())
		if err := validateFnFormat(method.Type, false); err != nil {
			return fmt.Errorf("failed to register activity method %v of %v: %e", methodName, structType.Name(), err)
		}

		structPrefix := options.Name
		registerName := methodName

		if options.EnableShortName {
			registerName = getShortFunctionName(methodName)
		}
		if len(structPrefix) > 0 {
			registerName = structPrefix + getShortFunctionName(methodName)
		}

		if !options.DisableAlreadyRegisteredCheck {
			if _, ok := r.getActivityNoLock(registerName); ok {
				return fmt.Errorf("activity type \"%v\" is already registered", registerName)
			}
		}
		r.activityFuncMap[registerName] = &activityExecutor{registerName, methodValue.Interface(), options, methodName}
		if len(structPrefix) > 0 || options.EnableShortName {
			r.activityAliasMap[methodName] = registerName
		}
		count++
	}

	if count == 0 {
		return fmt.Errorf("no activities (public methods) found in %v structure", structType.Name())
	}

	return nil
}

func getShortFunctionName(fnName string) string {
	elements := strings.Split(fnName, ".")
	return elements[len(elements)-1]
}

func (r *registry) getWorkflowAlias(fnName string) (string, bool) {
	r.Lock() // do not defer for Unlock to call next.getWorkflowAlias without lock
	alias, ok := r.workflowAliasMap[fnName]
	if !ok && r.next != nil {
		r.Unlock()
		return r.next.getWorkflowAlias(fnName)
	}
	r.Unlock()
	return alias, ok
}

func (r *registry) getWorkflowFn(fnName string) (interface{}, bool) {
	r.Lock() // do not defer for Unlock to call next.getWorkflowFn without lock
	wf, ok := r.workflowFuncMap[fnName]
	if !ok { // if exact match is not found, check for backwards compatible name without -fm suffix
		wf, ok = r.workflowFuncMap[strings.TrimSuffix(fnName, "-fm")]
	}
	if !ok && r.next != nil {
		r.Unlock()
		return r.next.getWorkflowFn(fnName)
	}
	r.Unlock()
	if ok {
		return wf.GetFunction(), ok
	}
	return nil, ok
}

func (r *registry) getWorkflowNoLock(registerName string) (interface{}, bool) {
	a, ok := r.workflowFuncMap[registerName]
	if !ok && r.next != nil {
		return r.next.getWorkflowNoLock(registerName)
	}
	if ok {
		return a.GetFunction(), ok
	}
	return nil, ok
}

func (r *registry) GetRegisteredWorkflowTypes() []string {
	r.Lock() // do not defer for Unlock to call next.getRegisteredWorkflowTypes without lock
	var result []string
	for t := range r.workflowFuncMap {
		result = append(result, t)
	}
	r.Unlock()
	if r.next != nil {
		nextTypes := r.next.GetRegisteredWorkflowTypes()
		result = append(result, nextTypes...)
	}
	return result
}

func (r *registry) getActivityAlias(fnName string) (string, bool) {
	r.Lock() // do not defer for Unlock to call next.getActivityAlias without lock
	alias, ok := r.activityAliasMap[fnName]
	if !ok && r.next != nil {
		r.Unlock()
		return r.next.getActivityAlias(fnName)
	}
	r.Unlock()
	return alias, ok
}

// Use in unit test only, otherwise deadlock will occur.
func (r *registry) addActivityWithLock(fnName string, a activity) {
	r.Lock()
	defer r.Unlock()
	r.activityFuncMap[fnName] = a
}

func (r *registry) GetActivity(fnName string) (activity, bool) {
	r.Lock() // do not defer for Unlock to call next.GetActivity without lock
	a, ok := r.activityFuncMap[fnName]
	if !ok { // if exact match is not found, check for backwards compatible name without -fm suffix
		a, ok = r.activityFuncMap[strings.TrimSuffix(fnName, "-fm")]
	}
	if !ok && r.next != nil {
		r.Unlock()
		return r.next.GetActivity(fnName)
	}
	r.Unlock()
	return a, ok
}

func (r *registry) getActivityNoLock(registerName string) (activity, bool) {
	a, ok := r.activityFuncMap[registerName]
	if !ok && r.next != nil {
		return r.next.getActivityNoLock(registerName)
	}
	return a, ok
}

func (r *registry) getRegisteredActivities() []activity {
	r.Lock() // do not defer for Unlock to call next.getRegisteredActivities without lock
	activities := make([]activity, 0, len(r.activityFuncMap))
	for _, a := range r.activityFuncMap {
		activities = append(activities, a)
	}
	r.Unlock()
	if r.next != nil {
		nextActivities := r.next.getRegisteredActivities()
		activities = append(activities, nextActivities...)
	}
	return activities
}

func (r *registry) getWorkflowDefinition(wt WorkflowType) (workflowDefinition, error) {
	lookup := getFunctionName(wt.Name)
	if alias, ok := r.getWorkflowAlias(lookup); ok {
		lookup = alias
	}
	wf, ok := r.getWorkflowFn(lookup)
	if !ok {
		supported := strings.Join(r.GetRegisteredWorkflowTypes(), ", ")
		return nil, fmt.Errorf(errMsgUnknownWorkflowType+": %v. Supported types: [%v]", lookup, supported)
	}
	wd := &workflowExecutor{workflowType: lookup, fn: wf}
	return newSyncWorkflowDefinition(wd), nil
}
