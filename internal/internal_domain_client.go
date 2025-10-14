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

	"github.com/uber-go/tally"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

// Assert that structs do indeed implement the interfaces
var _ DomainClient = (*domainClient)(nil)

// domainClient is the client for managing domains.
type domainClient struct {
	workflowService workflowserviceclient.Interface
	metricsScope    tally.Scope
	identity        string
	featureFlags    FeatureFlags
}

// Register a domain with cadence server
// The errors it can throw:
//   - DomainAlreadyExistsError
//   - BadRequestError
//   - InternalServiceError
func (dc *domainClient) Register(ctx context.Context, request *s.RegisterDomainRequest) error {
	return retryWhileTransientError(
		ctx,
		func() error {
			tchCtx, cancel, opt := newChannelContext(ctx, dc.featureFlags)
			defer cancel()
			return dc.workflowService.RegisterDomain(tchCtx, request, opt...)
		},
	)
}

// Describe a domain. The domain has 3 part of information
// DomainInfo - Which has Name, Status, Description, Owner Email
// DomainConfiguration - Configuration like Workflow Execution Retention Period In Days, Whether to emit metrics.
// ReplicationConfiguration - replication config like clusters and active cluster name
// The errors it can throw:
//   - EntityNotExistsError
//   - BadRequestError
//   - InternalServiceError
func (dc *domainClient) Describe(ctx context.Context, name string) (*s.DescribeDomainResponse, error) {
	request := &s.DescribeDomainRequest{
		Name: common.StringPtr(name),
	}

	var response *s.DescribeDomainResponse
	err := retryWhileTransientError(
		ctx,
		func() error {
			tchCtx, cancel, opt := newChannelContext(ctx, dc.featureFlags)
			defer cancel()
			var err error
			response, err = dc.workflowService.DescribeDomain(tchCtx, request, opt...)
			return err
		},
	)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Update a domain.
// The errors it can throw:
//   - EntityNotExistsError
//   - BadRequestError
//   - InternalServiceError
func (dc *domainClient) Update(ctx context.Context, request *s.UpdateDomainRequest) error {
	return retryWhileTransientError(
		ctx,
		func() error {
			tchCtx, cancel, opt := newChannelContext(ctx, dc.featureFlags)
			defer cancel()
			_, err := dc.workflowService.UpdateDomain(tchCtx, request, opt...)
			return err
		},
	)
}

// Failover a domain to another cluster.
// The errors it can throw:
//   - EntityNotExistsError
//   - BadRequestError
//   - InternalServiceError
func (dc *domainClient) Failover(ctx context.Context, request *s.FailoverDomainRequest) error {
	return retryWhileTransientError(
		ctx,
		func() error {
			tchCtx, cancel, opt := newChannelContext(ctx, dc.featureFlags)
			defer cancel()
			_, err := dc.workflowService.FailoverDomain(tchCtx, request, opt...)
			return err
		},
	)
}
