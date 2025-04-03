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
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/yarpc"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/metrics"
)

type domainClientTestData struct {
	dc                  DomainClient
	mockWorkflowService *workflowservicetest.MockClient
}

func newDomainClientTestData(t *testing.T) *domainClientTestData {
	var td domainClientTestData
	ctrl := gomock.NewController(t)

	metricsScope := metrics.NewTaggedScope(nil)

	td.mockWorkflowService = workflowservicetest.NewMockClient(ctrl)
	td.dc = NewDomainClient(
		td.mockWorkflowService,
		&ClientOptions{
			MetricsScope: metricsScope,
			Identity:     identity,
		},
	)

	return &td
}

func TestRegisterDomain(t *testing.T) {
	testcases := []struct {
		name     string
		rpcError error
	}{
		{
			name:     "success",
			rpcError: nil,
		},
		{
			name:     "failure",
			rpcError: &s.AccessDeniedError{},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newDomainClientTestData(t)
			request := &s.RegisterDomainRequest{
				Name: common.StringPtr(testDomain),
			}

			td.mockWorkflowService.EXPECT().
				RegisterDomain(gomock.Any(), request, gomock.Any()).
				Return(tt.rpcError)

			err := td.dc.Register(context.Background(), request)
			assert.Equal(t, tt.rpcError, err, "error should be returned as-is")
		})
	}
}

func TestDescribeDomain(t *testing.T) {
	testcases := []struct {
		name     string
		rpcError error
		response *s.DescribeDomainResponse
	}{
		{
			name:     "success",
			rpcError: nil,
			response: &s.DescribeDomainResponse{},
		},
		{
			name:     "failure",
			rpcError: &s.AccessDeniedError{},
			response: nil,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newDomainClientTestData(t)

			td.mockWorkflowService.EXPECT().
				DescribeDomain(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(
					func(ctx context.Context, req *s.DescribeDomainRequest, opts ...yarpc.CallOption) {
						// request also can contain other fields like UUID
						//   so compare a single, most important field explicitly
						assert.Equal(t, testDomain, *req.Name)
					},
				).
				Return(tt.response, tt.rpcError)

			r, err := td.dc.Describe(context.Background(), testDomain)
			assert.Equal(t, tt.rpcError, err, "error should be returned as-is")
			assert.Equal(t, tt.response, r)
		})
	}
}

func TestUpdateDomain(t *testing.T) {
	testcases := []struct {
		name     string
		rpcError error
	}{
		{
			name:     "success",
			rpcError: nil,
		},
		{
			name:     "failure",
			rpcError: &s.AccessDeniedError{},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newDomainClientTestData(t)
			request := &s.UpdateDomainRequest{
				Name: common.StringPtr(testDomain),
				UpdatedInfo: &s.UpdateDomainInfo{
					Description: common.StringPtr("domain for unit test"),
				},
			}

			td.mockWorkflowService.EXPECT().
				UpdateDomain(gomock.Any(), request, gomock.Any()).
				Return(nil, tt.rpcError)

			err := td.dc.Update(context.Background(), request)
			assert.Equal(t, tt.rpcError, err, "error should be returned as-is")
		})
	}
}
