package internal

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

func TestSemaphoreClientCreate(t *testing.T) {
	request := &CreateSemaphoreRequest{SemaphoreName: "semaphore", Capacity: 2}
	expectedRequest := &s.CreateSemaphoreRequest{
		Domain:         common.StringPtr("domain"),
		SemaphoreName:  common.StringPtr("semaphore"),
		Capacity:       common.Int32Ptr(2),
		BucketCapacity: common.Int32Ptr(0),
	}

	testcases := []struct {
		name     string
		response *s.CreateSemaphoreResponse
		rpcError error
		want     *Semaphore
	}{
		{
			name: "success",
			response: &s.CreateSemaphoreResponse{Semaphore: &s.Semaphore{
				SemaphoreName:  common.StringPtr("semaphore"),
				Capacity:       common.Int32Ptr(2),
				BucketCapacity: common.Int32Ptr(100),
			}},
			want: &Semaphore{SemaphoreName: "semaphore", Capacity: 2, BucketCapacity: 100},
		},
		{
			name:     "failure",
			rpcError: &s.BadRequestError{},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			service := workflowservicetest.NewMockClient(gomock.NewController(t))
			service.EXPECT().CreateSemaphore(gomock.Any(), expectedRequest, gomock.Any()).Return(tt.response, tt.rpcError)
			sc := NewClient(service, "domain", nil).SemaphoreClient()

			got, err := sc.Create(context.Background(), request)
			assert.Equal(t, tt.rpcError, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSemaphoreClientCreateNilRequest(t *testing.T) {
	sc := NewClient(workflowservicetest.NewMockClient(gomock.NewController(t)), "domain", nil).SemaphoreClient()

	_, err := sc.Create(context.Background(), nil)
	assert.ErrorIs(t, err, ErrRequestRequired)
}
