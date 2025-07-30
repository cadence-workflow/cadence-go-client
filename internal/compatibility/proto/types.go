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

package proto

import (
	"go.uber.org/cadence/.gen/go/shared"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

func Payload(data []byte) *apiv1.Payload {
	if data == nil {
		return nil
	}
	return &apiv1.Payload{
		Data: data,
	}
}

func Failure(reason *string, details []byte) *apiv1.Failure {
	if reason == nil {
		return nil
	}
	return &apiv1.Failure{
		Reason:  *reason,
		Details: details,
	}
}

func WorkflowExecution(t *shared.WorkflowExecution) *apiv1.WorkflowExecution {
	if t == nil {
		return nil
	}
	if t.WorkflowId == nil && t.RunId == nil {
		return nil
	}
	return &apiv1.WorkflowExecution{
		WorkflowId: t.GetWorkflowId(),
		RunId:      t.GetRunId(),
	}
}

func WorkflowRunPair(workflowID, runID string) *apiv1.WorkflowExecution {
	if workflowID == "" && runID == "" {
		return nil
	}
	return &apiv1.WorkflowExecution{
		WorkflowId: workflowID,
		RunId:      runID,
	}
}

func ActivityType(t *shared.ActivityType) *apiv1.ActivityType {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityType{
		Name: t.GetName(),
	}
}

func AutoConfigHint(t *shared.AutoConfigHint) *apiv1.AutoConfigHint {
	if t == nil {
		return nil
	}
	return &apiv1.AutoConfigHint{
		EnableAutoConfig:   t.GetEnableAutoConfig(),
		PollerWaitTimeInMs: t.GetPollerWaitTimeInMs(),
	}
}

func WorkflowType(t *shared.WorkflowType) *apiv1.WorkflowType {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowType{
		Name: t.GetName(),
	}
}

func TaskList(t *shared.TaskList) *apiv1.TaskList {
	if t == nil {
		return nil
	}
	return &apiv1.TaskList{
		Name: t.GetName(),
		Kind: TaskListKind(t.Kind),
	}
}

func TaskListMetadata(t *shared.TaskListMetadata) *apiv1.TaskListMetadata {
	if t == nil {
		return nil
	}
	return &apiv1.TaskListMetadata{
		MaxTasksPerSecond: fromDoubleValue(t.MaxTasksPerSecond),
	}
}

func RetryPolicy(t *shared.RetryPolicy) *apiv1.RetryPolicy {
	if t == nil {
		return nil
	}
	return &apiv1.RetryPolicy{
		InitialInterval:          secondsToDuration(t.InitialIntervalInSeconds),
		BackoffCoefficient:       t.GetBackoffCoefficient(),
		MaximumInterval:          secondsToDuration(t.MaximumIntervalInSeconds),
		MaximumAttempts:          t.GetMaximumAttempts(),
		NonRetryableErrorReasons: t.NonRetriableErrorReasons,
		ExpirationInterval:       secondsToDuration(t.ExpirationIntervalInSeconds),
	}
}

func Header(t *shared.Header) *apiv1.Header {
	if t == nil {
		return nil
	}
	return &apiv1.Header{
		Fields: PayloadMap(t.Fields),
	}
}

func Memo(t *shared.Memo) *apiv1.Memo {
	if t == nil {
		return nil
	}
	return &apiv1.Memo{
		Fields: PayloadMap(t.Fields),
	}
}

func SearchAttributes(t *shared.SearchAttributes) *apiv1.SearchAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.SearchAttributes{
		IndexedFields: PayloadMap(t.IndexedFields),
	}
}

func BadBinaries(t *shared.BadBinaries) *apiv1.BadBinaries {
	if t == nil {
		return nil
	}
	return &apiv1.BadBinaries{
		Binaries: BadBinaryInfoMap(t.Binaries),
	}
}

func BadBinaryInfo(t *shared.BadBinaryInfo) *apiv1.BadBinaryInfo {
	if t == nil {
		return nil
	}
	return &apiv1.BadBinaryInfo{
		Reason:      t.GetReason(),
		Operator:    t.GetOperator(),
		CreatedTime: unixNanoToTime(t.CreatedTimeNano),
	}
}

func ClusterReplicationConfiguration(t *shared.ClusterReplicationConfiguration) *apiv1.ClusterReplicationConfiguration {
	if t == nil {
		return nil
	}
	return &apiv1.ClusterReplicationConfiguration{
		ClusterName: t.GetClusterName(),
	}
}

func WorkflowQuery(t *shared.WorkflowQuery) *apiv1.WorkflowQuery {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowQuery{
		QueryType: t.GetQueryType(),
		QueryArgs: Payload(t.QueryArgs),
	}
}

func WorkflowQueryResult(t *shared.WorkflowQueryResult) *apiv1.WorkflowQueryResult {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowQueryResult{
		ResultType:   QueryResultType(t.ResultType),
		Answer:       Payload(t.Answer),
		ErrorMessage: t.GetErrorMessage(),
	}
}

func StickyExecutionAttributes(t *shared.StickyExecutionAttributes) *apiv1.StickyExecutionAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.StickyExecutionAttributes{
		WorkerTaskList:         TaskList(t.WorkerTaskList),
		ScheduleToStartTimeout: secondsToDuration(t.ScheduleToStartTimeoutSeconds),
	}
}

func WorkerVersionInfo(t *shared.WorkerVersionInfo) *apiv1.WorkerVersionInfo {
	if t == nil {
		return nil
	}
	return &apiv1.WorkerVersionInfo{
		Impl:           t.GetImpl(),
		FeatureVersion: t.GetFeatureVersion(),
	}
}

func StartTimeFilter(t *shared.StartTimeFilter) *apiv1.StartTimeFilter {
	if t == nil {
		return nil
	}
	return &apiv1.StartTimeFilter{
		EarliestTime: unixNanoToTime(t.EarliestTime),
		LatestTime:   unixNanoToTime(t.LatestTime),
	}
}

func WorkflowExecutionFilter(t *shared.WorkflowExecutionFilter) *apiv1.WorkflowExecutionFilter {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionFilter{
		WorkflowId: t.GetWorkflowId(),
		RunId:      t.GetRunId(),
	}
}

func WorkflowTypeFilter(t *shared.WorkflowTypeFilter) *apiv1.WorkflowTypeFilter {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowTypeFilter{
		Name: t.GetName(),
	}
}

func StatusFilter(t *shared.WorkflowExecutionCloseStatus) *apiv1.StatusFilter {
	if t == nil {
		return nil
	}
	return &apiv1.StatusFilter{
		Status: WorkflowExecutionCloseStatus(t),
	}
}

func PayloadMap(t map[string][]byte) map[string]*apiv1.Payload {
	if t == nil {
		return nil
	}
	v := make(map[string]*apiv1.Payload, len(t))
	for key := range t {
		v[key] = Payload(t[key])
	}
	return v
}

func BadBinaryInfoMap(t map[string]*shared.BadBinaryInfo) map[string]*apiv1.BadBinaryInfo {
	if t == nil {
		return nil
	}
	v := make(map[string]*apiv1.BadBinaryInfo, len(t))
	for key := range t {
		v[key] = BadBinaryInfo(t[key])
	}
	return v
}

func ClusterReplicationConfigurationArray(t []*shared.ClusterReplicationConfiguration) []*apiv1.ClusterReplicationConfiguration {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.ClusterReplicationConfiguration, len(t))
	for i := range t {
		v[i] = ClusterReplicationConfiguration(t[i])
	}
	return v
}

func WorkflowQueryResultMap(t map[string]*shared.WorkflowQueryResult) map[string]*apiv1.WorkflowQueryResult {
	if t == nil {
		return nil
	}
	v := make(map[string]*apiv1.WorkflowQueryResult, len(t))
	for key := range t {
		v[key] = WorkflowQueryResult(t[key])
	}
	return v
}

func DataBlob(t *shared.DataBlob) *apiv1.DataBlob {
	if t == nil {
		return nil
	}
	return &apiv1.DataBlob{
		EncodingType: EncodingType(t.EncodingType),
		Data:         t.Data,
	}
}

func ExternalExecutionInfo(we *shared.WorkflowExecution, initiatedID *int64) *apiv1.ExternalExecutionInfo {
	if we == nil && initiatedID == nil {
		return nil
	}
	if we == nil || initiatedID == nil {
		panic("either all or none external execution info fields must be set")
	}
	return &apiv1.ExternalExecutionInfo{
		WorkflowExecution: WorkflowExecution(we),
		InitiatedId:       *initiatedID,
	}
}

func ResetPoints(t *shared.ResetPoints) *apiv1.ResetPoints {
	if t == nil {
		return nil
	}
	return &apiv1.ResetPoints{
		Points: ResetPointInfoArray(t.Points),
	}
}

func ResetPointInfo(t *shared.ResetPointInfo) *apiv1.ResetPointInfo {
	if t == nil {
		return nil
	}
	return &apiv1.ResetPointInfo{
		BinaryChecksum:           t.GetBinaryChecksum(),
		RunId:                    t.GetRunId(),
		FirstDecisionCompletedId: t.GetFirstDecisionCompletedId(),
		CreatedTime:              unixNanoToTime(t.CreatedTimeNano),
		ExpiringTime:             unixNanoToTime(t.ExpiringTimeNano),
		Resettable:               t.GetResettable(),
	}
}

func PollerInfo(t *shared.PollerInfo) *apiv1.PollerInfo {
	if t == nil {
		return nil
	}
	return &apiv1.PollerInfo{
		LastAccessTime: unixNanoToTime(t.LastAccessTime),
		Identity:       t.GetIdentity(),
		RatePerSecond:  t.GetRatePerSecond(),
	}
}

func TaskListStatus(t *shared.TaskListStatus) *apiv1.TaskListStatus {
	if t == nil {
		return nil
	}
	return &apiv1.TaskListStatus{
		BacklogCountHint: t.GetBacklogCountHint(),
		ReadLevel:        t.GetReadLevel(),
		AckLevel:         t.GetAckLevel(),
		RatePerSecond:    t.GetRatePerSecond(),
		TaskIdBlock:      TaskIDBlock(t.TaskIDBlock),
	}
}

func TaskIDBlock(t *shared.TaskIDBlock) *apiv1.TaskIDBlock {
	if t == nil {
		return nil
	}
	return &apiv1.TaskIDBlock{
		StartId: t.GetStartID(),
		EndId:   t.GetEndID(),
	}
}

func WorkflowExecutionConfiguration(t *shared.WorkflowExecutionConfiguration) *apiv1.WorkflowExecutionConfiguration {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionConfiguration{
		TaskList:                     TaskList(t.TaskList),
		ExecutionStartToCloseTimeout: secondsToDuration(t.ExecutionStartToCloseTimeoutSeconds),
		TaskStartToCloseTimeout:      secondsToDuration(t.TaskStartToCloseTimeoutSeconds),
	}
}

func WorkflowExecutionInfo(t *shared.WorkflowExecutionInfo) *apiv1.WorkflowExecutionInfo {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionInfo{
		WorkflowExecution:   WorkflowExecution(t.Execution),
		Type:                WorkflowType(t.Type),
		StartTime:           unixNanoToTime(t.StartTime),
		CloseTime:           unixNanoToTime(t.CloseTime),
		CloseStatus:         WorkflowExecutionCloseStatus(t.CloseStatus),
		HistoryLength:       t.GetHistoryLength(),
		ParentExecutionInfo: ParentExecutionInfo2(t.ParentDomainId, t.ParentExecution),
		ExecutionTime:       unixNanoToTime(t.ExecutionTime),
		Memo:                Memo(t.Memo),
		SearchAttributes:    SearchAttributes(t.SearchAttributes),
		AutoResetPoints:     ResetPoints(t.AutoResetPoints),
		TaskList:            t.GetTaskList(),
		IsCron:              t.GetIsCron(),
	}
}

func ParentExecutionInfo(domainID, domainName *string, we *shared.WorkflowExecution, initiatedID *int64) *apiv1.ParentExecutionInfo {
	if domainID == nil && domainName == nil && we == nil && initiatedID == nil {
		return nil
	}
	if domainName == nil || we == nil || initiatedID == nil {
		panic("either all or none parent execution info must be set")
	}

	// Domain ID was added to unify parent execution info in several places.
	// However it may not be present:
	// - on older histories
	// - if conversion involves thrift data types
	// Fallback to empty string in those cases
	parentDomainID := ""
	if domainID != nil {
		parentDomainID = *domainID
	}

	return &apiv1.ParentExecutionInfo{
		DomainId:          parentDomainID,
		DomainName:        *domainName,
		WorkflowExecution: WorkflowExecution(we),
		InitiatedId:       *initiatedID,
	}
}

func ParentExecutionInfo2(domainID *string, we *shared.WorkflowExecution) *apiv1.ParentExecutionInfo {
	if domainID == nil && we == nil {
		return nil
	}

	return &apiv1.ParentExecutionInfo{
		DomainId:          *domainID,
		WorkflowExecution: WorkflowExecution(we),
	}
}

func PendingActivityInfo(t *shared.PendingActivityInfo) *apiv1.PendingActivityInfo {
	if t == nil {
		return nil
	}
	return &apiv1.PendingActivityInfo{
		ActivityId:         t.GetActivityID(),
		ActivityType:       ActivityType(t.ActivityType),
		State:              PendingActivityState(t.State),
		HeartbeatDetails:   Payload(t.HeartbeatDetails),
		LastHeartbeatTime:  unixNanoToTime(t.LastHeartbeatTimestamp),
		LastStartedTime:    unixNanoToTime(t.LastStartedTimestamp),
		Attempt:            t.GetAttempt(),
		MaximumAttempts:    t.GetMaximumAttempts(),
		ScheduledTime:      unixNanoToTime(t.ScheduledTimestamp),
		ExpirationTime:     unixNanoToTime(t.ExpirationTimestamp),
		LastFailure:        Failure(t.LastFailureReason, t.LastFailureDetails),
		LastWorkerIdentity: t.GetLastWorkerIdentity(),
	}
}

func PendingChildExecutionInfo(t *shared.PendingChildExecutionInfo) *apiv1.PendingChildExecutionInfo {
	if t == nil {
		return nil
	}
	return &apiv1.PendingChildExecutionInfo{
		WorkflowExecution: WorkflowRunPair(t.GetWorkflowID(), t.GetRunID()),
		WorkflowTypeName:  t.GetWorkflowTypName(),
		InitiatedId:       t.GetInitiatedID(),
		ParentClosePolicy: ParentClosePolicy(t.ParentClosePolicy),
	}
}

func PendingDecisionInfo(t *shared.PendingDecisionInfo) *apiv1.PendingDecisionInfo {
	if t == nil {
		return nil
	}
	return &apiv1.PendingDecisionInfo{
		State:                 PendingDecisionState(t.State),
		ScheduledTime:         unixNanoToTime(t.ScheduledTimestamp),
		StartedTime:           unixNanoToTime(t.StartedTimestamp),
		Attempt:               int32(t.GetAttempt()),
		OriginalScheduledTime: unixNanoToTime(t.OriginalScheduledTimestamp),
	}
}

func ActivityLocalDispatchInfo(t *shared.ActivityLocalDispatchInfo) *apiv1.ActivityLocalDispatchInfo {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityLocalDispatchInfo{
		ActivityId:                 t.GetActivityId(),
		ScheduledTime:              unixNanoToTime(t.ScheduledTimestamp),
		StartedTime:                unixNanoToTime(t.StartedTimestamp),
		ScheduledTimeOfThisAttempt: unixNanoToTime(t.ScheduledTimestampOfThisAttempt),
		TaskToken:                  t.TaskToken,
	}
}

func SupportedClientVersions(t *shared.SupportedClientVersions) *apiv1.SupportedClientVersions {
	if t == nil {
		return nil
	}
	return &apiv1.SupportedClientVersions{
		GoSdk:   t.GetGoSdk(),
		JavaSdk: t.GetJavaSdk(),
	}
}

func DescribeDomainResponseDomain(t *shared.DescribeDomainResponse) *apiv1.Domain {
	if t == nil {
		return nil
	}
	domain := apiv1.Domain{
		FailoverVersion: t.GetFailoverVersion(),
		IsGlobalDomain:  t.GetIsGlobalDomain(),
	}
	if info := t.DomainInfo; info != nil {
		domain.Id = info.GetUUID()
		domain.Name = info.GetName()
		domain.Status = DomainStatus(info.Status)
		domain.Description = info.GetDescription()
		domain.OwnerEmail = info.GetOwnerEmail()
		domain.Data = info.Data
	}
	if config := t.Configuration; config != nil {
		domain.WorkflowExecutionRetentionPeriod = daysToDuration(config.WorkflowExecutionRetentionPeriodInDays)
		domain.BadBinaries = BadBinaries(config.BadBinaries)
		domain.HistoryArchivalStatus = ArchivalStatus(config.HistoryArchivalStatus)
		domain.HistoryArchivalUri = config.GetHistoryArchivalURI()
		domain.VisibilityArchivalStatus = ArchivalStatus(config.VisibilityArchivalStatus)
		domain.VisibilityArchivalUri = config.GetVisibilityArchivalURI()
	}
	if repl := t.ReplicationConfiguration; repl != nil {
		domain.ActiveClusterName = repl.GetActiveClusterName()
		domain.Clusters = ClusterReplicationConfigurationArray(repl.Clusters)
		domain.ActiveClusters = ActiveClusters(repl.ActiveClusters)
	}
	return &domain
}

func TaskListPartitionMetadata(t *shared.TaskListPartitionMetadata) *apiv1.TaskListPartitionMetadata {
	if t == nil {
		return nil
	}
	return &apiv1.TaskListPartitionMetadata{
		Key:           t.GetKey(),
		OwnerHostName: t.GetOwnerHostName(),
	}
}

func QueryRejected(t *shared.QueryRejected) *apiv1.QueryRejected {
	if t == nil {
		return nil
	}
	return &apiv1.QueryRejected{
		CloseStatus: WorkflowExecutionCloseStatus(t.CloseStatus),
	}
}

func PollerInfoArray(t []*shared.PollerInfo) []*apiv1.PollerInfo {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.PollerInfo, len(t))
	for i := range t {
		v[i] = PollerInfo(t[i])
	}
	return v
}

func ResetPointInfoArray(t []*shared.ResetPointInfo) []*apiv1.ResetPointInfo {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.ResetPointInfo, len(t))
	for i := range t {
		v[i] = ResetPointInfo(t[i])
	}
	return v
}

func PendingActivityInfoArray(t []*shared.PendingActivityInfo) []*apiv1.PendingActivityInfo {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.PendingActivityInfo, len(t))
	for i := range t {
		v[i] = PendingActivityInfo(t[i])
	}
	return v
}

func PendingChildExecutionInfoArray(t []*shared.PendingChildExecutionInfo) []*apiv1.PendingChildExecutionInfo {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.PendingChildExecutionInfo, len(t))
	for i := range t {
		v[i] = PendingChildExecutionInfo(t[i])
	}
	return v
}

func IndexedValueTypeMap(t map[string]shared.IndexedValueType) map[string]apiv1.IndexedValueType {
	if t == nil {
		return nil
	}
	v := make(map[string]apiv1.IndexedValueType, len(t))
	for key := range t {
		v[key] = IndexedValueType(t[key])
	}
	return v
}

func DataBlobArray(t []*shared.DataBlob) []*apiv1.DataBlob {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.DataBlob, len(t))
	for i := range t {
		v[i] = DataBlob(t[i])
	}
	return v
}

func WorkflowExecutionInfoArray(t []*shared.WorkflowExecutionInfo) []*apiv1.WorkflowExecutionInfo {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.WorkflowExecutionInfo, len(t))
	for i := range t {
		v[i] = WorkflowExecutionInfo(t[i])
	}
	return v
}

func DescribeDomainResponseArray(t []*shared.DescribeDomainResponse) []*apiv1.Domain {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.Domain, len(t))
	for i := range t {
		v[i] = DescribeDomainResponseDomain(t[i])
	}
	return v
}

func TaskListPartitionMetadataArray(t []*shared.TaskListPartitionMetadata) []*apiv1.TaskListPartitionMetadata {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.TaskListPartitionMetadata, len(t))
	for i := range t {
		v[i] = TaskListPartitionMetadata(t[i])
	}
	return v
}

func WorkflowQueryMap(t map[string]*shared.WorkflowQuery) map[string]*apiv1.WorkflowQuery {
	if t == nil {
		return nil
	}
	v := make(map[string]*apiv1.WorkflowQuery, len(t))
	for key := range t {
		v[key] = WorkflowQuery(t[key])
	}
	return v
}

func ActivityLocalDispatchInfoMap(t map[string]*shared.ActivityLocalDispatchInfo) map[string]*apiv1.ActivityLocalDispatchInfo {
	if t == nil {
		return nil
	}
	v := make(map[string]*apiv1.ActivityLocalDispatchInfo, len(t))
	for key := range t {
		v[key] = ActivityLocalDispatchInfo(t[key])
	}
	return v
}

func ActiveClusters(ac *shared.ActiveClusters) *apiv1.ActiveClusters {
	if ac == nil {
		return nil
	}

	regToCl := make(map[string]*apiv1.ActiveClusterInfo)
	for reg, clusterInfo := range ac.ActiveClustersByRegion {
		regToCl[reg] = &apiv1.ActiveClusterInfo{
			ActiveClusterName: clusterInfo.GetActiveClusterName(),
			FailoverVersion:   clusterInfo.GetFailoverVersion(),
		}
	}

	return &apiv1.ActiveClusters{
		RegionToCluster: regToCl,
	}
}
