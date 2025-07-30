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

package thrift

import (
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

func Payload(p *apiv1.Payload) []byte {
	if p == nil {
		return nil
	}
	if p.Data == nil {
		// protoPayload will not generate this case
		// however, Data field will be dropped by the encoding if it's empty
		// and receiver side will see nil for the Data field
		// since we already know p is not nil, Data field must be an empty byte array
		return []byte{}
	}
	return p.Data
}

func Domain(domain string) *string {
	if domain == "" {
		return nil
	}
	return &domain
}

func FailureReason(failure *apiv1.Failure) *string {
	if failure == nil {
		return nil
	}
	return &failure.Reason
}

func FailureDetails(failure *apiv1.Failure) []byte {
	if failure == nil {
		return nil
	}
	return failure.Details
}

func WorkflowExecution(t *apiv1.WorkflowExecution) *shared.WorkflowExecution {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecution{
		WorkflowId: &t.WorkflowId,
		RunId:      &t.RunId,
	}
}

func WorkflowID(t *apiv1.WorkflowExecution) *string {
	if t == nil {
		return nil
	}
	return &t.WorkflowId
}

func RunID(t *apiv1.WorkflowExecution) *string {
	if t == nil {
		return nil
	}
	return &t.RunId
}

func ActivityType(t *apiv1.ActivityType) *shared.ActivityType {
	if t == nil {
		return nil
	}
	return &shared.ActivityType{
		Name: &t.Name,
	}
}

func AutoConfigHint(t *apiv1.AutoConfigHint) *shared.AutoConfigHint {
	if t == nil {
		return nil
	}
	return &shared.AutoConfigHint{
		EnableAutoConfig:   &t.EnableAutoConfig,
		PollerWaitTimeInMs: &t.PollerWaitTimeInMs,
	}
}

func WorkflowType(t *apiv1.WorkflowType) *shared.WorkflowType {
	if t == nil {
		return nil
	}
	return &shared.WorkflowType{
		Name: &t.Name,
	}
}

func TaskList(t *apiv1.TaskList) *shared.TaskList {
	if t == nil {
		return nil
	}
	return &shared.TaskList{
		Name: &t.Name,
		Kind: TaskListKind(t.Kind),
	}
}

func TaskListMetadata(t *apiv1.TaskListMetadata) *shared.TaskListMetadata {
	if t == nil {
		return nil
	}
	return &shared.TaskListMetadata{
		MaxTasksPerSecond: toDoubleValue(t.MaxTasksPerSecond),
	}
}

func RetryPolicy(t *apiv1.RetryPolicy) *shared.RetryPolicy {
	if t == nil {
		return nil
	}
	return &shared.RetryPolicy{
		InitialIntervalInSeconds:    durationToSeconds(t.InitialInterval),
		BackoffCoefficient:          &t.BackoffCoefficient,
		MaximumIntervalInSeconds:    durationToSeconds(t.MaximumInterval),
		MaximumAttempts:             &t.MaximumAttempts,
		NonRetriableErrorReasons:    t.NonRetryableErrorReasons,
		ExpirationIntervalInSeconds: durationToSeconds(t.ExpirationInterval),
	}
}

func Header(t *apiv1.Header) *shared.Header {
	if t == nil {
		return nil
	}
	return &shared.Header{
		Fields: PayloadMap(t.Fields),
	}
}

func Memo(t *apiv1.Memo) *shared.Memo {
	if t == nil {
		return nil
	}
	return &shared.Memo{
		Fields: PayloadMap(t.Fields),
	}
}

func SearchAttributes(t *apiv1.SearchAttributes) *shared.SearchAttributes {
	if t == nil {
		return nil
	}
	return &shared.SearchAttributes{
		IndexedFields: PayloadMap(t.IndexedFields),
	}
}

func BadBinaries(t *apiv1.BadBinaries) *shared.BadBinaries {
	if t == nil {
		return nil
	}
	return &shared.BadBinaries{
		Binaries: BadBinaryInfoMap(t.Binaries),
	}
}

func BadBinaryInfo(t *apiv1.BadBinaryInfo) *shared.BadBinaryInfo {
	if t == nil {
		return nil
	}
	return &shared.BadBinaryInfo{
		Reason:          &t.Reason,
		Operator:        &t.Operator,
		CreatedTimeNano: timeToUnixNano(t.CreatedTime),
	}
}

func ClusterReplicationConfiguration(t *apiv1.ClusterReplicationConfiguration) *shared.ClusterReplicationConfiguration {
	if t == nil {
		return nil
	}
	return &shared.ClusterReplicationConfiguration{
		ClusterName: &t.ClusterName,
	}
}

func WorkflowQuery(t *apiv1.WorkflowQuery) *shared.WorkflowQuery {
	if t == nil {
		return nil
	}
	return &shared.WorkflowQuery{
		QueryType: &t.QueryType,
		QueryArgs: Payload(t.QueryArgs),
	}
}

func WorkflowQueryResult(t *apiv1.WorkflowQueryResult) *shared.WorkflowQueryResult {
	if t == nil {
		return nil
	}
	return &shared.WorkflowQueryResult{
		ResultType:   QueryResultType(t.ResultType),
		Answer:       Payload(t.Answer),
		ErrorMessage: &t.ErrorMessage,
	}
}

func StickyExecutionAttributes(t *apiv1.StickyExecutionAttributes) *shared.StickyExecutionAttributes {
	if t == nil {
		return nil
	}
	return &shared.StickyExecutionAttributes{
		WorkerTaskList:                TaskList(t.WorkerTaskList),
		ScheduleToStartTimeoutSeconds: durationToSeconds(t.ScheduleToStartTimeout),
	}
}

func WorkerVersionInfo(t *apiv1.WorkerVersionInfo) *shared.WorkerVersionInfo {
	if t == nil {
		return nil
	}
	return &shared.WorkerVersionInfo{
		Impl:           &t.Impl,
		FeatureVersion: &t.FeatureVersion,
	}
}

func StartTimeFilter(t *apiv1.StartTimeFilter) *shared.StartTimeFilter {
	if t == nil {
		return nil
	}
	return &shared.StartTimeFilter{
		EarliestTime: timeToUnixNano(t.EarliestTime),
		LatestTime:   timeToUnixNano(t.LatestTime),
	}
}

func WorkflowExecutionFilter(t *apiv1.WorkflowExecutionFilter) *shared.WorkflowExecutionFilter {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionFilter{
		WorkflowId: &t.WorkflowId,
		RunId:      &t.RunId,
	}
}

func WorkflowTypeFilter(t *apiv1.WorkflowTypeFilter) *shared.WorkflowTypeFilter {
	if t == nil {
		return nil
	}
	return &shared.WorkflowTypeFilter{
		Name: &t.Name,
	}
}

func StatusFilter(t *apiv1.StatusFilter) *shared.WorkflowExecutionCloseStatus {
	if t == nil {
		return nil
	}
	return WorkflowExecutionCloseStatus(t.Status)
}

func PayloadMap(t map[string]*apiv1.Payload) map[string][]byte {
	if t == nil {
		return nil
	}
	v := make(map[string][]byte, len(t))
	for key := range t {
		v[key] = Payload(t[key])
	}
	return v
}

func BadBinaryInfoMap(t map[string]*apiv1.BadBinaryInfo) map[string]*shared.BadBinaryInfo {
	if t == nil {
		return nil
	}
	v := make(map[string]*shared.BadBinaryInfo, len(t))
	for key := range t {
		v[key] = BadBinaryInfo(t[key])
	}
	return v
}

func ClusterReplicationConfigurationArray(t []*apiv1.ClusterReplicationConfiguration) []*shared.ClusterReplicationConfiguration {
	if t == nil {
		return nil
	}
	v := make([]*shared.ClusterReplicationConfiguration, len(t))
	for i := range t {
		v[i] = ClusterReplicationConfiguration(t[i])
	}
	return v
}

func WorkflowQueryResultMap(t map[string]*apiv1.WorkflowQueryResult) map[string]*shared.WorkflowQueryResult {
	if t == nil {
		return nil
	}
	v := make(map[string]*shared.WorkflowQueryResult, len(t))
	for key := range t {
		v[key] = WorkflowQueryResult(t[key])
	}
	return v
}

func DataBlob(t *apiv1.DataBlob) *shared.DataBlob {
	if t == nil {
		return nil
	}
	return &shared.DataBlob{
		EncodingType: EncodingType(t.EncodingType),
		Data:         t.Data,
	}
}

func ExternalInitiatedID(t *apiv1.ExternalExecutionInfo) *int64 {
	if t == nil {
		return nil
	}
	return &t.InitiatedId
}

func ExternalWorkflowExecution(t *apiv1.ExternalExecutionInfo) *shared.WorkflowExecution {
	if t == nil {
		return nil
	}
	return WorkflowExecution(t.WorkflowExecution)
}

func ResetPoints(t *apiv1.ResetPoints) *shared.ResetPoints {
	if t == nil {
		return nil
	}
	return &shared.ResetPoints{
		Points: ResetPointInfoArray(t.Points),
	}
}

func ResetPointInfo(t *apiv1.ResetPointInfo) *shared.ResetPointInfo {
	if t == nil {
		return nil
	}
	return &shared.ResetPointInfo{
		BinaryChecksum:           &t.BinaryChecksum,
		RunId:                    &t.RunId,
		FirstDecisionCompletedId: &t.FirstDecisionCompletedId,
		CreatedTimeNano:          timeToUnixNano(t.CreatedTime),
		ExpiringTimeNano:         timeToUnixNano(t.ExpiringTime),
		Resettable:               &t.Resettable,
	}
}

func PollerInfo(t *apiv1.PollerInfo) *shared.PollerInfo {
	if t == nil {
		return nil
	}
	return &shared.PollerInfo{
		LastAccessTime: timeToUnixNano(t.LastAccessTime),
		Identity:       &t.Identity,
		RatePerSecond:  &t.RatePerSecond,
	}
}

func TaskListStatus(t *apiv1.TaskListStatus) *shared.TaskListStatus {
	if t == nil {
		return nil
	}
	return &shared.TaskListStatus{
		BacklogCountHint: &t.BacklogCountHint,
		ReadLevel:        &t.ReadLevel,
		AckLevel:         &t.AckLevel,
		RatePerSecond:    &t.RatePerSecond,
		TaskIDBlock:      TaskIDBlock(t.TaskIdBlock),
	}
}

func TaskIDBlock(t *apiv1.TaskIDBlock) *shared.TaskIDBlock {
	if t == nil {
		return nil
	}
	return &shared.TaskIDBlock{
		StartID: &t.StartId,
		EndID:   &t.EndId,
	}
}

func WorkflowExecutionConfiguration(t *apiv1.WorkflowExecutionConfiguration) *shared.WorkflowExecutionConfiguration {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionConfiguration{
		TaskList:                            TaskList(t.TaskList),
		ExecutionStartToCloseTimeoutSeconds: durationToSeconds(t.ExecutionStartToCloseTimeout),
		TaskStartToCloseTimeoutSeconds:      durationToSeconds(t.TaskStartToCloseTimeout),
	}
}

func WorkflowExecutionInfo(t *apiv1.WorkflowExecutionInfo) *shared.WorkflowExecutionInfo {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionInfo{
		Execution:        WorkflowExecution(t.WorkflowExecution),
		Type:             WorkflowType(t.Type),
		StartTime:        timeToUnixNano(t.StartTime),
		CloseTime:        timeToUnixNano(t.CloseTime),
		CloseStatus:      WorkflowExecutionCloseStatus(t.CloseStatus),
		HistoryLength:    &t.HistoryLength,
		ParentDomainId:   ParentDomainID(t.ParentExecutionInfo),
		ParentExecution:  ParentWorkflowExecution(t.ParentExecutionInfo),
		ExecutionTime:    timeToUnixNano(t.ExecutionTime),
		Memo:             Memo(t.Memo),
		SearchAttributes: SearchAttributes(t.SearchAttributes),
		AutoResetPoints:  ResetPoints(t.AutoResetPoints),
		TaskList:         &t.TaskList,
		IsCron:           &t.IsCron,
	}
}

func ParentDomainID(pei *apiv1.ParentExecutionInfo) *string {
	if pei == nil {
		return nil
	}
	return &pei.DomainId
}

func ParentDomainName(pei *apiv1.ParentExecutionInfo) *string {
	if pei == nil {
		return nil
	}
	return &pei.DomainName
}

func ParentInitiatedID(pei *apiv1.ParentExecutionInfo) *int64 {
	if pei == nil {
		return nil
	}
	return &pei.InitiatedId
}

func ParentWorkflowExecution(pei *apiv1.ParentExecutionInfo) *shared.WorkflowExecution {
	if pei == nil {
		return nil
	}
	return WorkflowExecution(pei.WorkflowExecution)
}

func PendingActivityInfo(t *apiv1.PendingActivityInfo) *shared.PendingActivityInfo {
	if t == nil {
		return nil
	}
	return &shared.PendingActivityInfo{
		ActivityID:             &t.ActivityId,
		ActivityType:           ActivityType(t.ActivityType),
		State:                  PendingActivityState(t.State),
		HeartbeatDetails:       Payload(t.HeartbeatDetails),
		LastHeartbeatTimestamp: timeToUnixNano(t.LastHeartbeatTime),
		LastStartedTimestamp:   timeToUnixNano(t.LastStartedTime),
		Attempt:                &t.Attempt,
		MaximumAttempts:        &t.MaximumAttempts,
		ScheduledTimestamp:     timeToUnixNano(t.ScheduledTime),
		ExpirationTimestamp:    timeToUnixNano(t.ExpirationTime),
		LastFailureReason:      FailureReason(t.LastFailure),
		LastFailureDetails:     FailureDetails(t.LastFailure),
		LastWorkerIdentity:     &t.LastWorkerIdentity,
	}
}

func PendingChildExecutionInfo(t *apiv1.PendingChildExecutionInfo) *shared.PendingChildExecutionInfo {
	if t == nil {
		return nil
	}
	return &shared.PendingChildExecutionInfo{
		WorkflowID:        WorkflowID(t.WorkflowExecution),
		RunID:             RunID(t.WorkflowExecution),
		WorkflowTypName:   &t.WorkflowTypeName,
		InitiatedID:       &t.InitiatedId,
		ParentClosePolicy: ParentClosePolicy(t.ParentClosePolicy),
	}
}

func PendingDecisionInfo(t *apiv1.PendingDecisionInfo) *shared.PendingDecisionInfo {
	if t == nil {
		return nil
	}
	return &shared.PendingDecisionInfo{
		State:                      PendingDecisionState(t.State),
		ScheduledTimestamp:         timeToUnixNano(t.ScheduledTime),
		StartedTimestamp:           timeToUnixNano(t.StartedTime),
		Attempt:                    common.Int64Ptr(int64(t.Attempt)),
		OriginalScheduledTimestamp: timeToUnixNano(t.OriginalScheduledTime),
	}
}

func ActivityLocalDispatchInfo(t *apiv1.ActivityLocalDispatchInfo) *shared.ActivityLocalDispatchInfo {
	if t == nil {
		return nil
	}
	return &shared.ActivityLocalDispatchInfo{
		ActivityId:                      &t.ActivityId,
		ScheduledTimestamp:              timeToUnixNano(t.ScheduledTime),
		StartedTimestamp:                timeToUnixNano(t.StartedTime),
		ScheduledTimestampOfThisAttempt: timeToUnixNano(t.ScheduledTimeOfThisAttempt),
		TaskToken:                       t.TaskToken,
	}
}

func SupportedClientVersions(t *apiv1.SupportedClientVersions) *shared.SupportedClientVersions {
	if t == nil {
		return nil
	}
	return &shared.SupportedClientVersions{
		GoSdk:   &t.GoSdk,
		JavaSdk: &t.JavaSdk,
	}
}

func DescribeDomainResponseDomain(t *apiv1.Domain) *shared.DescribeDomainResponse {
	if t == nil {
		return nil
	}
	return &shared.DescribeDomainResponse{
		DomainInfo: &shared.DomainInfo{
			Name:        &t.Name,
			Status:      DomainStatus(t.Status),
			Description: &t.Description,
			OwnerEmail:  &t.OwnerEmail,
			Data:        t.Data,
			UUID:        &t.Id,
		},
		Configuration: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: durationToDays(t.WorkflowExecutionRetentionPeriod),
			EmitMetric:                             boolPtr(true),
			BadBinaries:                            BadBinaries(t.BadBinaries),
			HistoryArchivalStatus:                  ArchivalStatus(t.HistoryArchivalStatus),
			HistoryArchivalURI:                     &t.HistoryArchivalUri,
			VisibilityArchivalStatus:               ArchivalStatus(t.VisibilityArchivalStatus),
			VisibilityArchivalURI:                  &t.VisibilityArchivalUri,
		},
		ReplicationConfiguration: &shared.DomainReplicationConfiguration{
			ActiveClusterName: &t.ActiveClusterName,
			Clusters:          ClusterReplicationConfigurationArray(t.Clusters),
			ActiveClusters:    ActiveClusters(t.ActiveClusters),
		},
		FailoverVersion: &t.FailoverVersion,
		IsGlobalDomain:  &t.IsGlobalDomain,
	}
}

func TaskListPartitionMetadata(t *apiv1.TaskListPartitionMetadata) *shared.TaskListPartitionMetadata {
	if t == nil {
		return nil
	}
	return &shared.TaskListPartitionMetadata{
		Key:           &t.Key,
		OwnerHostName: &t.OwnerHostName,
	}
}

func QueryRejected(t *apiv1.QueryRejected) *shared.QueryRejected {
	if t == nil {
		return nil
	}
	return &shared.QueryRejected{
		CloseStatus: WorkflowExecutionCloseStatus(t.CloseStatus),
	}
}

func PollerInfoArray(t []*apiv1.PollerInfo) []*shared.PollerInfo {
	if t == nil {
		return nil
	}
	v := make([]*shared.PollerInfo, len(t))
	for i := range t {
		v[i] = PollerInfo(t[i])
	}
	return v
}

func ResetPointInfoArray(t []*apiv1.ResetPointInfo) []*shared.ResetPointInfo {
	if t == nil {
		return nil
	}
	v := make([]*shared.ResetPointInfo, len(t))
	for i := range t {
		v[i] = ResetPointInfo(t[i])
	}
	return v
}

func PendingActivityInfoArray(t []*apiv1.PendingActivityInfo) []*shared.PendingActivityInfo {
	if t == nil {
		return nil
	}
	v := make([]*shared.PendingActivityInfo, len(t))
	for i := range t {
		v[i] = PendingActivityInfo(t[i])
	}
	return v
}

func PendingChildExecutionInfoArray(t []*apiv1.PendingChildExecutionInfo) []*shared.PendingChildExecutionInfo {
	if t == nil {
		return nil
	}
	v := make([]*shared.PendingChildExecutionInfo, len(t))
	for i := range t {
		v[i] = PendingChildExecutionInfo(t[i])
	}
	return v
}

func IndexedValueTypeMap(t map[string]apiv1.IndexedValueType) map[string]shared.IndexedValueType {
	if t == nil {
		return nil
	}
	v := make(map[string]shared.IndexedValueType, len(t))
	for key := range t {
		v[key] = IndexedValueType(t[key])
	}
	return v
}

func DataBlobArray(t []*apiv1.DataBlob) []*shared.DataBlob {
	if t == nil {
		return nil
	}
	v := make([]*shared.DataBlob, len(t))
	for i := range t {
		v[i] = DataBlob(t[i])
	}
	return v
}

func WorkflowExecutionInfoArray(t []*apiv1.WorkflowExecutionInfo) []*shared.WorkflowExecutionInfo {
	if t == nil {
		return nil
	}
	v := make([]*shared.WorkflowExecutionInfo, len(t))
	for i := range t {
		v[i] = WorkflowExecutionInfo(t[i])
	}
	return v
}

func DescribeDomainResponseArray(t []*apiv1.Domain) []*shared.DescribeDomainResponse {
	if t == nil {
		return nil
	}
	v := make([]*shared.DescribeDomainResponse, len(t))
	for i := range t {
		v[i] = DescribeDomainResponseDomain(t[i])
	}
	return v
}

func TaskListPartitionMetadataArray(t []*apiv1.TaskListPartitionMetadata) []*shared.TaskListPartitionMetadata {
	if t == nil {
		return nil
	}
	v := make([]*shared.TaskListPartitionMetadata, len(t))
	for i := range t {
		v[i] = TaskListPartitionMetadata(t[i])
	}
	return v
}

func WorkflowQueryMap(t map[string]*apiv1.WorkflowQuery) map[string]*shared.WorkflowQuery {
	if t == nil {
		return nil
	}
	v := make(map[string]*shared.WorkflowQuery, len(t))
	for key := range t {
		v[key] = WorkflowQuery(t[key])
	}
	return v
}

func ActivityLocalDispatchInfoMap(t map[string]*apiv1.ActivityLocalDispatchInfo) map[string]*shared.ActivityLocalDispatchInfo {
	if t == nil {
		return nil
	}
	v := make(map[string]*shared.ActivityLocalDispatchInfo, len(t))
	for key := range t {
		v[key] = ActivityLocalDispatchInfo(t[key])
	}
	return v
}

func ActiveClusters(ac *apiv1.ActiveClusters) *shared.ActiveClusters {
	if ac == nil {
		return nil
	}

	clByRegion := make(map[string]*shared.ActiveClusterInfo)
	for region, clInfo := range ac.RegionToCluster {
		clByRegion[region] = &shared.ActiveClusterInfo{
			ActiveClusterName: &clInfo.ActiveClusterName,
			FailoverVersion:   &clInfo.FailoverVersion,
		}
	}

	return &shared.ActiveClusters{
		ActiveClustersByRegion: clByRegion,
	}
}
