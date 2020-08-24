// // Copyright (c) 2017-2020 Uber Technologies Inc.
// //
// // Permission is hereby granted, free of charge, to any person obtaining a copy
// // of this software and associated documentation files (the "Software"), to deal
// // in the Software without restriction, including without limitation the rights
// // to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// // copies of the Software, and to permit persons to whom the Software is
// // furnished to do so, subject to the following conditions:
// //
// // The above copyright notice and this permission notice shall be included in
// // all copies or substantial portions of the Software.
// //
// // THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// // IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// // FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// // AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// // LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// // OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// // THE SOFTWARE.
//
//

// Code generated by MockGen. DO NOT EDIT.
// Source: go.uber.org/cadence/internal (interfaces: DomainClient)

// Package mockgen is a generated GoMock package.
package mockgen

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	shared "go.uber.org/cadence/.gen/go/shared"
	reflect "reflect"
)

// MockDomainClient is a mock of DomainClient interface
type MockDomainClient struct {
	ctrl     *gomock.Controller
	recorder *MockDomainClientMockRecorder
}

// MockDomainClientMockRecorder is the mock recorder for MockDomainClient
type MockDomainClientMockRecorder struct {
	mock *MockDomainClient
}

// NewMockDomainClient creates a new mock instance
func NewMockDomainClient(ctrl *gomock.Controller) *MockDomainClient {
	mock := &MockDomainClient{ctrl: ctrl}
	mock.recorder = &MockDomainClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDomainClient) EXPECT() *MockDomainClientMockRecorder {
	return m.recorder
}

// Describe mocks base method
func (m *MockDomainClient) Describe(arg0 context.Context, arg1 string) (*shared.DescribeDomainResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Describe", arg0, arg1)
	ret0, _ := ret[0].(*shared.DescribeDomainResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Describe indicates an expected call of Describe
func (mr *MockDomainClientMockRecorder) Describe(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Describe", reflect.TypeOf((*MockDomainClient)(nil).Describe), arg0, arg1)
}

// Register mocks base method
func (m *MockDomainClient) Register(arg0 context.Context, arg1 *shared.RegisterDomainRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register
func (mr *MockDomainClientMockRecorder) Register(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockDomainClient)(nil).Register), arg0, arg1)
}

// Update mocks base method
func (m *MockDomainClient) Update(arg0 context.Context, arg1 *shared.UpdateDomainRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockDomainClientMockRecorder) Update(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockDomainClient)(nil).Update), arg0, arg1)
}
