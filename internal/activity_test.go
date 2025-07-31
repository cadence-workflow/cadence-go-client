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
	"context"
	"testing"

	"go.uber.org/cadence/internal/common/testlogger"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/yarpc"
	"go.uber.org/zap"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

const (
	testWorkflowType = "test-workflow-type"
	testActivityType = "test-activity-type"
)

type activityTestSuite struct {
	suite.Suite
	mockCtrl *gomock.Controller
	service  *workflowservicetest.MockClient
	logger   *zap.Logger
}

func TestActivityTestSuite(t *testing.T) {
	s := new(activityTestSuite)
	suite.Run(t, s)
}

func (s *activityTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.service = workflowservicetest.NewMockClient(s.mockCtrl)
	s.logger = testlogger.NewZap(s.T())
}

func (s *activityTestSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

func (s *activityTestSuite) TestActivityHeartbeat() {
	ctx, cancel := context.WithCancel(context.Background())
	invoker := newServiceInvoker([]byte("task-token"), "identity", s.service, cancel, 1, make(chan struct{}), FeatureFlags{}, s.logger, testWorkflowType, testActivityType)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{serviceInvoker: invoker})

	s.service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions()...).
		Return(&shared.RecordActivityTaskHeartbeatResponse{}, nil).Times(1)

	RecordActivityHeartbeat(ctx, "testDetails")
}

func (s *activityTestSuite) TestActivityHeartbeat_InternalError() {
	ctx, cancel := context.WithCancel(context.Background())
	invoker := newServiceInvoker([]byte("task-token"), "identity", s.service, cancel, 1, make(chan struct{}), FeatureFlags{}, s.logger, testWorkflowType, testActivityType)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		serviceInvoker: invoker,
		logger:         getTestLogger(s.T())})

	s.service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions()...).
		Return(nil, &shared.InternalServiceError{}).
		Do(func(ctx context.Context, request *shared.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) {
			s.T().Log("MOCK RecordActivityTaskHeartbeat executed")
		}).AnyTimes()

	RecordActivityHeartbeat(ctx, "testDetails")
}

func (s *activityTestSuite) TestActivityHeartbeat_CancelRequested() {
	ctx, cancel := context.WithCancel(context.Background())
	invoker := newServiceInvoker([]byte("task-token"), "identity", s.service, cancel, 1, make(chan struct{}), FeatureFlags{}, s.logger, testWorkflowType, testActivityType)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		serviceInvoker: invoker,
		logger:         getTestLogger(s.T())})

	s.service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions()...).
		Return(&shared.RecordActivityTaskHeartbeatResponse{CancelRequested: common.BoolPtr(true)}, nil).Times(1)

	RecordActivityHeartbeat(ctx, "testDetails")
	<-ctx.Done()
	require.Equal(s.T(), ctx.Err(), context.Canceled)
}

func (s *activityTestSuite) TestActivityHeartbeat_EntityNotExist() {
	ctx, cancel := context.WithCancel(context.Background())
	invoker := newServiceInvoker([]byte("task-token"), "identity", s.service, cancel, 1, make(chan struct{}), FeatureFlags{}, s.logger, testWorkflowType, testActivityType)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		serviceInvoker: invoker,
		logger:         getTestLogger(s.T())})

	s.service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions()...).
		Return(&shared.RecordActivityTaskHeartbeatResponse{}, &shared.EntityNotExistsError{}).Times(1)

	RecordActivityHeartbeat(ctx, "testDetails")
	<-ctx.Done()
	require.Equal(s.T(), ctx.Err(), context.Canceled)
}

func (s *activityTestSuite) TestActivityHeartbeat_SuppressContinousInvokes() {
	ctx, cancel := context.WithCancel(context.Background())
	invoker := newServiceInvoker([]byte("task-token"), "identity", s.service, cancel, 2, make(chan struct{}), FeatureFlags{}, s.logger, testWorkflowType, testActivityType)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		serviceInvoker: invoker,
		logger:         getTestLogger(s.T())})

	// Multiple calls but only one call is made.
	s.service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions()...).
		Return(&shared.RecordActivityTaskHeartbeatResponse{}, nil).Times(1)
	RecordActivityHeartbeat(ctx, "testDetails")
	RecordActivityHeartbeat(ctx, "testDetails")
	RecordActivityHeartbeat(ctx, "testDetails")
	invoker.Close(false)

	// No HB timeout configured.
	service2 := workflowservicetest.NewMockClient(s.mockCtrl)
	invoker2 := newServiceInvoker([]byte("task-token"), "identity", service2, cancel, 0, make(chan struct{}), FeatureFlags{}, s.logger, testWorkflowType, testActivityType)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		serviceInvoker: invoker2,
		logger:         getTestLogger(s.T())})
	service2.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions()...).
		Return(&shared.RecordActivityTaskHeartbeatResponse{}, nil).Times(1)
	RecordActivityHeartbeat(ctx, "testDetails")
	RecordActivityHeartbeat(ctx, "testDetails")
	invoker2.Close(false)

	// simulate batch picks before expiry.
	waitCh := make(chan struct{})
	service3 := workflowservicetest.NewMockClient(s.mockCtrl)
	invoker3 := newServiceInvoker([]byte("task-token"), "identity", service3, cancel, 2, make(chan struct{}), FeatureFlags{}, s.logger, testWorkflowType, testActivityType)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		serviceInvoker: invoker3,
		logger:         getTestLogger(s.T())})
	service3.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions()...).
		Return(&shared.RecordActivityTaskHeartbeatResponse{}, nil).Times(1)

	service3.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions()...).
		Return(&shared.RecordActivityTaskHeartbeatResponse{}, nil).
		Do(func(ctx context.Context, request *shared.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) {
			ev := newEncodedValues(request.Details, nil)
			var progress string
			err := ev.Get(&progress)
			if err != nil {
				panic(err)
			}
			require.Equal(s.T(), "testDetails-expected", progress)
			waitCh <- struct{}{}
		}).Times(1)

	RecordActivityHeartbeat(ctx, "testDetails")
	RecordActivityHeartbeat(ctx, "testDetails2")
	RecordActivityHeartbeat(ctx, "testDetails3")
	RecordActivityHeartbeat(ctx, "testDetails-expected")
	<-waitCh
	invoker3.Close(false)

	// simulate batch picks before expiry, with out any progress specified.
	waitCh2 := make(chan struct{})
	service4 := workflowservicetest.NewMockClient(s.mockCtrl)
	invoker4 := newServiceInvoker([]byte("task-token"), "identity", service4, cancel, 2, make(chan struct{}), FeatureFlags{}, s.logger, testWorkflowType, testActivityType)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		serviceInvoker: invoker4,
		logger:         getTestLogger(s.T())})
	service4.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions()...).
		Return(&shared.RecordActivityTaskHeartbeatResponse{}, nil).Times(1)
	service4.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions()...).
		Return(&shared.RecordActivityTaskHeartbeatResponse{}, nil).
		Do(func(ctx context.Context, request *shared.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) {
			require.Nil(s.T(), request.Details)
			waitCh2 <- struct{}{}
		}).Times(1)

	RecordActivityHeartbeat(ctx, nil)
	RecordActivityHeartbeat(ctx, nil)
	RecordActivityHeartbeat(ctx, nil)
	RecordActivityHeartbeat(ctx, nil)
	<-waitCh2
	invoker4.Close(false)
}

func (s *activityTestSuite) TestActivityHeartbeat_WorkerStop() {
	ctx, cancel := context.WithCancel(context.Background())
	workerStopChannel := make(chan struct{})
	invoker := newServiceInvoker([]byte("task-token"), "identity", s.service, cancel, 5, workerStopChannel, FeatureFlags{}, s.logger, testWorkflowType, testActivityType)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{serviceInvoker: invoker})

	heartBeatDetail := "testDetails"
	waitCh := make(chan struct{}, 1)
	waitCh <- struct{}{}
	waitC2 := make(chan struct{}, 1)
	s.service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions()...).
		Return(&shared.RecordActivityTaskHeartbeatResponse{}, nil).
		Do(func(ctx context.Context, request *shared.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) {
			if _, ok := <-waitCh; ok {
				close(waitCh)
				return
			}
			close(waitC2)
		}).Times(2)
	RecordActivityHeartbeat(ctx, heartBeatDetail)
	RecordActivityHeartbeat(ctx, "testDetails")
	close(workerStopChannel)
	<-waitC2
}

func (s *activityTestSuite) TestGetWorkerStopChannel() {
	ch := make(chan struct{}, 1)
	ctx := context.WithValue(context.Background(), activityEnvContextKey, &activityEnvironment{workerStopChannel: ch})
	channel := GetWorkerStopChannel(ctx)
	s.NotNil(channel)
}

func (s *activityTestSuite) TestHasActivityInfo() {
	// Test context without activity info
	ctx := context.Background()
	s.False(HasActivityInfo(ctx))

	// Test context with activity info
	activityEnv := &activityEnvironment{
		activityID:   "test-activity-id",
		activityType: ActivityType{Name: "test-activity-type"},
	}
	ctxWithActivity := context.WithValue(ctx, activityEnvContextKey, activityEnv)
	s.True(HasActivityInfo(ctxWithActivity))

	// Test context with nil activity env
	ctxWithNilActivity := context.WithValue(ctx, activityEnvContextKey, nil)
	s.False(HasActivityInfo(ctxWithNilActivity))

	// Test context with other values in context
	ctxWithOtherValue := context.WithValue(ctx, activityOptionsContextKey, "other-value")
	s.False(HasActivityInfo(ctxWithOtherValue))
}
