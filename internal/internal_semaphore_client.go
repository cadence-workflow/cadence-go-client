//go:generate mockery --srcpkg . --name SemaphoreClient --output ../mocks --with-expecter

package internal

import (
	"context"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	shared "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

// SemaphoreClient is the client for managing Cadence semaphores within a domain.
type SemaphoreClient interface {
	// Create creates a semaphore. If the semaphore already exists with the same capacity and bucket
	// capacity, it returns the existing one instead of an error.
	// The errors it can return:
	//  - BadRequestError
	//  - EntityNotExistError
	//  - InternalServiceError
	Create(ctx context.Context, request *CreateSemaphoreRequest) (*Semaphore, error)
}

var _ SemaphoreClient = (*semaphoreClient)(nil)

type semaphoreClient struct {
	workflowService workflowserviceclient.Interface
	domain          string
	featureFlags    FeatureFlags
}

// SemaphoreClient implements Client. It returns a SemaphoreClient scoped to this client's domain and connection.
func (wc *workflowClient) SemaphoreClient() SemaphoreClient {
	return &semaphoreClient{
		workflowService: wc.workflowService,
		domain:          wc.domain,
		featureFlags:    wc.featureFlags,
	}
}

func (sc *semaphoreClient) Create(ctx context.Context, request *CreateSemaphoreRequest) (*Semaphore, error) {
	if request == nil {
		return nil, ErrRequestRequired
	}
	createRequest := &shared.CreateSemaphoreRequest{
		Domain:         common.StringPtr(sc.domain),
		SemaphoreName:  common.StringPtr(request.SemaphoreName),
		Capacity:       common.Int32Ptr(request.Capacity),
		BucketCapacity: common.Int32Ptr(request.BucketCapacity),
	}
	var semaphore *Semaphore
	err := retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		resp, err := sc.workflowService.CreateSemaphore(tchCtx, createRequest, opt...)
		if err == nil {
			semaphore = semaphoreFromThrift(resp.GetSemaphore())
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return semaphore, nil
}

func semaphoreFromThrift(t *shared.Semaphore) *Semaphore {
	if t == nil {
		return nil
	}
	return &Semaphore{
		SemaphoreName:  t.GetSemaphoreName(),
		Capacity:       t.GetCapacity(),
		BucketCapacity: t.GetBucketCapacity(),
	}
}
