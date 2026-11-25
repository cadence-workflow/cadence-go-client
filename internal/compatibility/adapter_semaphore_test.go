package compatibility

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/yarpc"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

type semaphoreStub struct {
	request *apiv1.CreateSemaphoreRequest
}

func (s *semaphoreStub) CreateSemaphore(_ context.Context, request *apiv1.CreateSemaphoreRequest, _ ...yarpc.CallOption) (*apiv1.CreateSemaphoreResponse, error) {
	s.request = request
	return &apiv1.CreateSemaphoreResponse{Semaphore: &apiv1.Semaphore{
		SemaphoreName: request.SemaphoreName,
		Capacity:      request.Capacity,
	}}, nil
}

func TestAdapterCreateSemaphore(t *testing.T) {
	stub := &semaphoreStub{}
	a := thrift2protoAdapter{semaphore: stub}

	resp, err := a.CreateSemaphore(context.Background(), &shared.CreateSemaphoreRequest{
		Domain:        common.StringPtr("domain"),
		SemaphoreName: common.StringPtr("semaphore"),
		Capacity:      common.Int32Ptr(2),
	})
	require.NoError(t, err)
	assert.Equal(t, &apiv1.CreateSemaphoreRequest{Domain: "domain", SemaphoreName: "semaphore", Capacity: 2}, stub.request)
	assert.Equal(t, "semaphore", resp.Semaphore.GetSemaphoreName())
	assert.Equal(t, int32(2), resp.Semaphore.GetCapacity())
}

func TestAdapterCreateSemaphoreNotConfigured(t *testing.T) {
	a := thrift2protoAdapter{}

	_, err := a.CreateSemaphore(context.Background(), &shared.CreateSemaphoreRequest{})
	var badRequestErr *shared.BadRequestError
	assert.True(t, errors.As(err, &badRequestErr))
}
