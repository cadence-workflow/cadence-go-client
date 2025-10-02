// Copyright (c) 2021 Uber Technologies, Inc.
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

package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/uber/tchannel-go/thrift"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	"go.uber.org/cadence/.gen/go/shared"
)

type (
	serviceWrapperSuite struct {
		suite.Suite
		Service      *workflowservicetest.MockClient
		AuthProvider AuthorizationProvider
		controller   *gomock.Controller
	}
)

type jwtAuthCorrect struct{}

func (j *jwtAuthCorrect) GetAuthToken() ([]byte, error) {
	return []byte{}, nil
}

func newJWTAuthCorrect() AuthorizationProvider {
	return &jwtAuthCorrect{}
}

type jwtAuthIncorrect struct{}

func (j *jwtAuthIncorrect) GetAuthToken() ([]byte, error) {
	return []byte{}, fmt.Errorf("error")
}

func newJWTAuthIncorrect() AuthorizationProvider {
	return &jwtAuthIncorrect{}
}

func TestServiceWrapperSuite(t *testing.T) {
	suite.Run(t, new(serviceWrapperSuite))
}

func (s *serviceWrapperSuite) SetupTest() {
	s.controller = gomock.NewController(s.T())
	s.Service = workflowservicetest.NewMockClient(s.controller)
	s.AuthProvider = newJWTAuthCorrect()
}

func (s *serviceWrapperSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *serviceWrapperSuite) TestDeprecateDomainValidToken() {
	s.Service.EXPECT().DeprecateDomain(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.DeprecateDomain(ctx, &shared.DeprecateDomainRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestDeprecateDomainInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.DeprecateDomain(ctx, &shared.DeprecateDomainRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestListDomainsValidToken() {
	s.Service.EXPECT().ListDomains(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ListDomains(ctx, &shared.ListDomainsRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestListDomainsInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ListDomains(ctx, &shared.ListDomainsRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestDescribeDomainValidToken() {
	s.Service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.DescribeDomain(ctx, &shared.DescribeDomainRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestDescribeDomainInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.DescribeDomain(ctx, &shared.DescribeDomainRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestDescribeWorkflowExecutionValidToken() {
	s.Service.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.DescribeWorkflowExecution(ctx, &shared.DescribeWorkflowExecutionRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestDescribeWorkflowExecutionInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.DescribeWorkflowExecution(ctx, &shared.DescribeWorkflowExecutionRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestDiagnoseWorkflowExecutionValidToken() {
	s.Service.EXPECT().DiagnoseWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.DiagnoseWorkflowExecution(ctx, &shared.DiagnoseWorkflowExecutionRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestDiagnoseWorkflowExecutionInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.DiagnoseWorkflowExecution(ctx, &shared.DiagnoseWorkflowExecutionRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestGetWorkflowExecutionHistoryValidToken() {
	s.Service.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.GetWorkflowExecutionHistory(ctx, &shared.GetWorkflowExecutionHistoryRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestGetWorkflowExecutionHistoryInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.GetWorkflowExecutionHistory(ctx, &shared.GetWorkflowExecutionHistoryRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestListClosedWorkflowExecutionsValidToken() {
	s.Service.EXPECT().ListClosedWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ListClosedWorkflowExecutions(ctx, &shared.ListClosedWorkflowExecutionsRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestListClosedWorkflowExecutionsInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ListClosedWorkflowExecutions(ctx, &shared.ListClosedWorkflowExecutionsRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestListOpenWorkflowExecutionsValidToken() {
	s.Service.EXPECT().ListOpenWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ListOpenWorkflowExecutions(ctx, &shared.ListOpenWorkflowExecutionsRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestListOpenWorkflowExecutionsInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ListOpenWorkflowExecutions(ctx, &shared.ListOpenWorkflowExecutionsRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestListWorkflowExecutionsValidToken() {
	s.Service.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ListWorkflowExecutions(ctx, &shared.ListWorkflowExecutionsRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestListWorkflowExecutionsInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ListWorkflowExecutions(ctx, &shared.ListWorkflowExecutionsRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestListArchivedWorkflowExecutionsValidToken() {
	s.Service.EXPECT().ListArchivedWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ListArchivedWorkflowExecutions(ctx, &shared.ListArchivedWorkflowExecutionsRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestListArchivedWorkflowExecutionsInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ListArchivedWorkflowExecutions(ctx, &shared.ListArchivedWorkflowExecutionsRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestScanWorkflowExecutionsValidToken() {
	s.Service.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ScanWorkflowExecutions(ctx, &shared.ListWorkflowExecutionsRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestScanWorkflowExecutionsInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ScanWorkflowExecutions(ctx, &shared.ListWorkflowExecutionsRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestCountWorkflowExecutionsValidToken() {
	s.Service.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.CountWorkflowExecutions(ctx, &shared.CountWorkflowExecutionsRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestCountWorkflowExecutionsInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.CountWorkflowExecutions(ctx, &shared.CountWorkflowExecutionsRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestPollForActivityTaskValidToken() {
	s.Service.EXPECT().PollForActivityTask(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.PollForActivityTask(ctx, &shared.PollForActivityTaskRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestPollForActivityTaskInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.PollForActivityTask(ctx, &shared.PollForActivityTaskRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestPollForDecisionTaskValidToken() {
	s.Service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.PollForDecisionTask(ctx, &shared.PollForDecisionTaskRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestPollForDecisionTaskInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.PollForDecisionTask(ctx, &shared.PollForDecisionTaskRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestRecordActivityTaskHeartbeatValidToken() {
	s.Service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.RecordActivityTaskHeartbeat(ctx, &shared.RecordActivityTaskHeartbeatRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestRecordActivityTaskHeartbeatInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.RecordActivityTaskHeartbeat(ctx, &shared.RecordActivityTaskHeartbeatRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestRecordActivityTaskHeartbeatByIDValidToken() {
	s.Service.EXPECT().RecordActivityTaskHeartbeatByID(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.RecordActivityTaskHeartbeatByID(ctx, &shared.RecordActivityTaskHeartbeatByIDRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestRecordActivityTaskHeartbeatByIDInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.RecordActivityTaskHeartbeatByID(ctx, &shared.RecordActivityTaskHeartbeatByIDRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestRegisterDomainValidToken() {
	s.Service.EXPECT().RegisterDomain(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RegisterDomain(ctx, &shared.RegisterDomainRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestRegisterDomainInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RegisterDomain(ctx, &shared.RegisterDomainRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestRequestCancelWorkflowExecutionValidToken() {
	s.Service.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RequestCancelWorkflowExecution(ctx, &shared.RequestCancelWorkflowExecutionRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestRequestCancelWorkflowExecutionInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RequestCancelWorkflowExecution(ctx, &shared.RequestCancelWorkflowExecutionRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestRespondActivityTaskCanceledValidToken() {
	s.Service.EXPECT().RespondActivityTaskCanceled(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondActivityTaskCanceled(ctx, &shared.RespondActivityTaskCanceledRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestRespondActivityTaskCanceledInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondActivityTaskCanceled(ctx, &shared.RespondActivityTaskCanceledRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestRespondActivityTaskCompletedValidToken() {
	s.Service.EXPECT().RespondActivityTaskCompleted(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondActivityTaskCompleted(ctx, &shared.RespondActivityTaskCompletedRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestRespondActivityTaskCompletedInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondActivityTaskCompleted(ctx, &shared.RespondActivityTaskCompletedRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestRespondActivityTaskFailedValidToken() {
	s.Service.EXPECT().RespondActivityTaskFailed(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondActivityTaskFailed(ctx, &shared.RespondActivityTaskFailedRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestRespondActivityTaskFailedInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondActivityTaskFailed(ctx, &shared.RespondActivityTaskFailedRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestRespondActivityTaskCanceledByIDValidToken() {
	s.Service.EXPECT().RespondActivityTaskCanceledByID(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondActivityTaskCanceledByID(ctx, &shared.RespondActivityTaskCanceledByIDRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestRespondActivityTaskCanceledByIDInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondActivityTaskCanceledByID(ctx, &shared.RespondActivityTaskCanceledByIDRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestRespondActivityTaskCompletedByIDValidToken() {
	s.Service.EXPECT().RespondActivityTaskCompletedByID(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondActivityTaskCompletedByID(ctx, &shared.RespondActivityTaskCompletedByIDRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestRespondActivityTaskCompletedByIDInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondActivityTaskCompletedByID(ctx, &shared.RespondActivityTaskCompletedByIDRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestRespondActivityTaskFailedByIDValidToken() {
	s.Service.EXPECT().RespondActivityTaskFailedByID(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondActivityTaskFailedByID(ctx, &shared.RespondActivityTaskFailedByIDRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestRespondActivityTaskFailedByIDInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondActivityTaskFailedByID(ctx, &shared.RespondActivityTaskFailedByIDRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestRespondDecisionTaskCompletedValidToken() {
	s.Service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.RespondDecisionTaskCompleted(ctx, &shared.RespondDecisionTaskCompletedRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestRespondDecisionTaskCompletedInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.RespondDecisionTaskCompleted(ctx, &shared.RespondDecisionTaskCompletedRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestRespondDecisionTaskFailedValidToken() {
	s.Service.EXPECT().RespondDecisionTaskFailed(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondDecisionTaskFailed(ctx, &shared.RespondDecisionTaskFailedRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestRespondDecisionTaskFailedInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondDecisionTaskFailed(ctx, &shared.RespondDecisionTaskFailedRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestSignalWorkflowExecutionValidToken() {
	s.Service.EXPECT().SignalWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.SignalWorkflowExecution(ctx, &shared.SignalWorkflowExecutionRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestSignalWorkflowExecutionInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.SignalWorkflowExecution(ctx, &shared.SignalWorkflowExecutionRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestSignalWithStartWorkflowExecutionToken() {
	s.Service.EXPECT().SignalWithStartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.SignalWithStartWorkflowExecution(ctx, &shared.SignalWithStartWorkflowExecutionRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestSignalWithStartWorkflowExecutionInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.SignalWithStartWorkflowExecution(ctx, &shared.SignalWithStartWorkflowExecutionRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestSignalWithStartWorkflowExecutionAsyncToken() {
	s.Service.EXPECT().SignalWithStartWorkflowExecutionAsync(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.SignalWithStartWorkflowExecutionAsync(ctx, &shared.SignalWithStartWorkflowExecutionAsyncRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestSignalWithStartWorkflowExecutionAsyncInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.SignalWithStartWorkflowExecutionAsync(ctx, &shared.SignalWithStartWorkflowExecutionAsyncRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestStartWorkflowExecutionToken() {
	s.Service.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.StartWorkflowExecution(ctx, &shared.StartWorkflowExecutionRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestStartWorkflowExecutionInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.StartWorkflowExecution(ctx, &shared.StartWorkflowExecutionRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestStartWorkflowExecutionAsyncToken() {
	s.Service.EXPECT().StartWorkflowExecutionAsync(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.StartWorkflowExecutionAsync(ctx, &shared.StartWorkflowExecutionAsyncRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestStartWorkflowExecutionAsyncInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.StartWorkflowExecutionAsync(ctx, &shared.StartWorkflowExecutionAsyncRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestTerminateWorkflowExecutionToken() {
	s.Service.EXPECT().TerminateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.TerminateWorkflowExecution(ctx, &shared.TerminateWorkflowExecutionRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestTerminateWorkflowExecutionInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.TerminateWorkflowExecution(ctx, &shared.TerminateWorkflowExecutionRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestResetWorkflowExecutionValidToken() {
	s.Service.EXPECT().ResetWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ResetWorkflowExecution(ctx, &shared.ResetWorkflowExecutionRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestResetWorkflowExecutionInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ResetWorkflowExecution(ctx, &shared.ResetWorkflowExecutionRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestUpdateDomainValidToken() {
	s.Service.EXPECT().UpdateDomain(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.UpdateDomain(ctx, &shared.UpdateDomainRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestUpdateDomainInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.UpdateDomain(ctx, &shared.UpdateDomainRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestQueryWorkflowValidToken() {
	s.Service.EXPECT().QueryWorkflow(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.QueryWorkflow(ctx, &shared.QueryWorkflowRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestQueryWorkflowInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.QueryWorkflow(ctx, &shared.QueryWorkflowRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestResetStickyTaskListValidToken() {
	s.Service.EXPECT().ResetStickyTaskList(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ResetStickyTaskList(ctx, &shared.ResetStickyTaskListRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestResetStickyTaskListInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ResetStickyTaskList(ctx, &shared.ResetStickyTaskListRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestDescribeTaskListValidToken() {
	s.Service.EXPECT().DescribeTaskList(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.DescribeTaskList(ctx, &shared.DescribeTaskListRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestDescribeTaskListInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.DescribeTaskList(ctx, &shared.DescribeTaskListRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestRespondQueryTaskCompletedValidToken() {
	s.Service.EXPECT().RespondQueryTaskCompleted(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondQueryTaskCompleted(ctx, &shared.RespondQueryTaskCompletedRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestRespondQueryTaskCompletedInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.RespondQueryTaskCompleted(ctx, &shared.RespondQueryTaskCompletedRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestGetSearchAttributesValidToken() {
	s.Service.EXPECT().GetSearchAttributes(gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.GetSearchAttributes(ctx)
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestGetSearchAttributesInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.GetSearchAttributes(ctx)
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestListTaskListPartitionsValidToken() {
	s.Service.EXPECT().ListTaskListPartitions(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ListTaskListPartitions(ctx, &shared.ListTaskListPartitionsRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestListTaskListPartitionsInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.ListTaskListPartitions(ctx, &shared.ListTaskListPartitionsRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestGetClusterInfoValidToken() {
	s.Service.EXPECT().GetClusterInfo(gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.GetClusterInfo(ctx)
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestGetClusterInfoInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.GetClusterInfo(ctx)
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestDeleteDomainValidToken() {
	s.Service.EXPECT().DeleteDomain(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.DeleteDomain(ctx, &shared.DeleteDomainRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestDeleteDomainInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	err := sw.DeleteDomain(ctx, &shared.DeleteDomainRequest{})
	s.EqualError(err, "error")
}

func (s *serviceWrapperSuite) TestFailoverDomainValidToken() {
	s.Service.EXPECT().FailoverDomain(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.FailoverDomain(ctx, &shared.FailoverDomainRequest{})
	s.NoError(err)
}

func (s *serviceWrapperSuite) TestFailoverDomainInvalidToken() {
	s.AuthProvider = newJWTAuthIncorrect()
	sw := NewWorkflowServiceWrapper(s.Service, s.AuthProvider)
	ctx, _ := thrift.NewContext(time.Minute)
	_, err := sw.FailoverDomain(ctx, &shared.FailoverDomainRequest{})
	s.EqualError(err, "error")
}
