// Copyright (c) 2021 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package testdata

import (
	"time"

	gogo "github.com/gogo/protobuf/types"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

const (
	WorkflowID = "WorkflowID"
	RunID      = "RunID"
	RunID1     = "RunID1"
	RunID2     = "RunID2"
	RunID3     = "RunID3"

	ActivityID = "ActivityID"
	RequestID  = "RequestID"
	TimerID    = "TimerID"

	WorkflowTypeName     = "WorkflowTypeName"
	ActivityTypeName     = "ActivityTypeName"
	TaskListName         = "TaskListName"
	MarkerName           = "MarkerName"
	SignalName           = "SignalName"
	QueryType            = "QueryType"
	HostName             = "HostName"
	Identity             = "Identity"
	CronSchedule         = "CronSchedule"
	CronOverlapPolicy    = 0
	Checksum             = "Checksum"
	Reason               = "Reason"
	Cause                = "Cause"
	SecurityToken        = "SecurityToken"
	VisibilityQuery      = "VisibilityQuery"
	FeatureVersion       = "FeatureVersion"
	ClientImpl           = "ClientImpl"
	ClientLibraryVersion = "ClientLibraryVersion"
	SupportedVersions    = "SupportedVersions"

	Attempt           = 2
	PageSize          = 10
	HistoryLength     = 20
	BacklogCountHint  = 30
	AckLevel          = 1001
	ReadLevel         = 1002
	RatePerSecond     = 3.14
	TaskID            = 444
	ShardID           = 12345
	MessageID1        = 50001
	MessageID2        = 50002
	EventStoreVersion = 333

	EventID1 = int64(1)
	EventID2 = int64(2)
	EventID3 = int64(3)
	EventID4 = int64(4)

	Version1 = int64(11)
	Version2 = int64(22)
	Version3 = int64(33)
)

var (
	ts            = time.Now()
	Timestamp, _  = gogo.TimestampProto(ts)
	Timestamp1, _ = gogo.TimestampProto(ts.Add(1 * time.Second))
	Timestamp2, _ = gogo.TimestampProto(ts.Add(2 * time.Second))
	Timestamp3, _ = gogo.TimestampProto(ts.Add(3 * time.Second))
	Timestamp4, _ = gogo.TimestampProto(ts.Add(4 * time.Second))
	Timestamp5, _ = gogo.TimestampProto(ts.Add(5 * time.Second))
)

var (
	Duration1 = gogo.DurationProto(11 * time.Second)
	Duration2 = gogo.DurationProto(12 * time.Second)
	Duration3 = gogo.DurationProto(13 * time.Second)
	Duration4 = gogo.DurationProto(14 * time.Second)
)

var (
	Token1 = []byte{1, 0}
	Token2 = []byte{2, 0}
	Token3 = []byte{3, 0}
)

var (
	Payload1 = apiv1.Payload{Data: []byte{10, 0}}
	Payload2 = apiv1.Payload{Data: []byte{20, 0}}
	Payload3 = apiv1.Payload{Data: []byte{30, 0}}
)

var (
	ExecutionContext = []byte{110, 0}
	Control          = []byte{120, 0}
	NextPageToken    = []byte{130, 0}
	TaskToken        = []byte{140, 0}
	BranchToken      = []byte{150, 0}
	QueryArgs        = &apiv1.Payload{Data: []byte{9, 9, 9}}
)

var (
	FailureReason  = "FailureReason"
	FailureDetails = []byte{190, 0}
)

var (
	Failure = apiv1.Failure{
		Reason:  FailureReason,
		Details: FailureDetails,
	}
	WorkflowExecution = apiv1.WorkflowExecution{
		WorkflowId: WorkflowID,
		RunId:      RunID,
	}
	WorkflowType = apiv1.WorkflowType{
		Name: WorkflowTypeName,
	}
	ActivityType = apiv1.ActivityType{
		Name: ActivityTypeName,
	}
	TaskList = apiv1.TaskList{
		Name: TaskListName,
		Kind: apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
	}
	RetryPolicy = apiv1.RetryPolicy{
		InitialInterval:          Duration1,
		BackoffCoefficient:       1.1,
		MaximumInterval:          Duration2,
		MaximumAttempts:          3,
		NonRetryableErrorReasons: []string{"a", "b"},
		ExpirationInterval:       Duration3,
	}
	Header = apiv1.Header{
		Fields: map[string]*apiv1.Payload{
			"HeaderField1": &apiv1.Payload{Data: []byte{211, 0}},
			"HeaderField2": &apiv1.Payload{Data: []byte{212, 0}},
		},
	}
	Memo = apiv1.Memo{
		Fields: map[string]*apiv1.Payload{
			"MemoField1": &apiv1.Payload{Data: []byte{221, 0}},
			"MemoField2": &apiv1.Payload{Data: []byte{222, 0}},
		},
	}
	SearchAttributes = apiv1.SearchAttributes{
		IndexedFields: map[string]*apiv1.Payload{
			"IndexedField1": &apiv1.Payload{Data: []byte{231, 0}},
			"IndexedField2": &apiv1.Payload{Data: []byte{232, 0}},
		},
	}

	ActiveClusterSelectionPolicyRegionSticky = &apiv1.ActiveClusterSelectionPolicy{
		Strategy: apiv1.ActiveClusterSelectionStrategy_ACTIVE_CLUSTER_SELECTION_STRATEGY_REGION_STICKY,
		StrategyConfig: &apiv1.ActiveClusterSelectionPolicy_ActiveClusterStickyRegionConfig{
			ActiveClusterStickyRegionConfig: &apiv1.ActiveClusterStickyRegionConfig{
				StickyRegion: "us-east-1",
			},
		},
	}
	ActiveClusterSelectionPolicyExternalEntity = &apiv1.ActiveClusterSelectionPolicy{
		Strategy: apiv1.ActiveClusterSelectionStrategy_ACTIVE_CLUSTER_SELECTION_STRATEGY_EXTERNAL_ENTITY,
		StrategyConfig: &apiv1.ActiveClusterSelectionPolicy_ActiveClusterExternalEntityConfig{
			ActiveClusterExternalEntityConfig: &apiv1.ActiveClusterExternalEntityConfig{
				ExternalEntityType: "external-entity-type",
				ExternalEntityKey:  "external-entity-key",
			},
		},
	}

	PayloadMap = map[string]*apiv1.Payload{
		"Payload1": &Payload1,
		"Payload2": &Payload2,
	}
	ResetPointInfo = apiv1.ResetPointInfo{
		BinaryChecksum:           Checksum,
		RunId:                    RunID1,
		FirstDecisionCompletedId: EventID1,
		CreatedTime:              Timestamp1,
		ExpiringTime:             Timestamp2,
		Resettable:               true,
	}
	ResetPointInfoArray = []*apiv1.ResetPointInfo{
		&ResetPointInfo,
	}
	ResetPoints = apiv1.ResetPoints{
		Points: ResetPointInfoArray,
	}
	DataBlob = apiv1.DataBlob{
		EncodingType: EncodingType,
		Data:         []byte{7, 7, 7},
	}
	DataBlobArray = []*apiv1.DataBlob{
		&DataBlob,
	}
	StickyExecutionAttributes = apiv1.StickyExecutionAttributes{
		WorkerTaskList:         &TaskList,
		ScheduleToStartTimeout: Duration1,
	}
	ActivityLocalDispatchInfo = apiv1.ActivityLocalDispatchInfo{
		ActivityId:                 ActivityID,
		ScheduledTime:              Timestamp1,
		StartedTime:                Timestamp2,
		ScheduledTimeOfThisAttempt: Timestamp3,
		TaskToken:                  TaskToken,
	}
	ActivityLocalDispatchInfoMap = map[string]*apiv1.ActivityLocalDispatchInfo{
		ActivityID: &ActivityLocalDispatchInfo,
	}
	AutoConfigHint = apiv1.AutoConfigHint{
		EnableAutoConfig:   true,
		PollerWaitTimeInMs: 100,
	}
	TaskListMetadata = apiv1.TaskListMetadata{
		MaxTasksPerSecond: &gogo.DoubleValue{Value: RatePerSecond},
	}
	WorkerVersionInfo = apiv1.WorkerVersionInfo{
		Impl:           ClientImpl,
		FeatureVersion: FeatureVersion,
	}
	PollerInfo = apiv1.PollerInfo{
		LastAccessTime: Timestamp1,
		Identity:       Identity,
		RatePerSecond:  RatePerSecond,
	}
	PollerInfoArray = []*apiv1.PollerInfo{
		&PollerInfo,
	}
	TaskListStatus = apiv1.TaskListStatus{
		BacklogCountHint: BacklogCountHint,
		ReadLevel:        ReadLevel,
		AckLevel:         AckLevel,
		RatePerSecond:    RatePerSecond,
		TaskIdBlock:      &TaskIDBlock,
	}
	TaskIDBlock = apiv1.TaskIDBlock{
		StartId: 551,
		EndId:   559,
	}
	TaskListPartitionMetadata = apiv1.TaskListPartitionMetadata{
		Key:           "Key",
		OwnerHostName: "HostName",
	}
	TaskListPartitionMetadataArray = []*apiv1.TaskListPartitionMetadata{
		&TaskListPartitionMetadata,
	}
	SupportedClientVersions = apiv1.SupportedClientVersions{
		GoSdk:   "GoSDK",
		JavaSdk: "JavaSDK",
	}
	IndexedValueTypeMap = map[string]apiv1.IndexedValueType{
		"IndexedValueType1": IndexedValueType,
	}
	ParentExecutionInfo = apiv1.ParentExecutionInfo{
		DomainId:          DomainID,
		DomainName:        DomainName,
		WorkflowExecution: &WorkflowExecution,
		InitiatedId:       EventID1,
	}
	WorkflowExecutionFilter = apiv1.WorkflowExecutionFilter{
		WorkflowId: WorkflowID,
		RunId:      RunID,
	}
	WorkflowTypeFilter = apiv1.WorkflowTypeFilter{
		Name: WorkflowTypeName,
	}
	StartTimeFilter = apiv1.StartTimeFilter{
		EarliestTime: Timestamp1,
		LatestTime:   Timestamp2,
	}
	StatusFilter = apiv1.StatusFilter{
		Status: WorkflowExecutionCloseStatus,
	}
	WorkflowExecutionInfo = apiv1.WorkflowExecutionInfo{
		WorkflowExecution: &WorkflowExecution,
		Type:              &WorkflowType,
		StartTime:         Timestamp1,
		CloseTime:         Timestamp2,
		CloseStatus:       WorkflowExecutionCloseStatus,
		HistoryLength:     HistoryLength,
		ParentExecutionInfo: &apiv1.ParentExecutionInfo{
			DomainId:          DomainID,
			WorkflowExecution: &WorkflowExecution,
		},
		ExecutionTime:    Timestamp3,
		Memo:             &Memo,
		SearchAttributes: &SearchAttributes,
		AutoResetPoints:  &ResetPoints,
		TaskList:         TaskListName,
	}
	WorkflowExecutionInfoArray = []*apiv1.WorkflowExecutionInfo{&WorkflowExecutionInfo}

	WorkflowQuery = apiv1.WorkflowQuery{
		QueryType: QueryType,
		QueryArgs: QueryArgs,
	}
	WorkflowQueryMap = map[string]*apiv1.WorkflowQuery{
		"WorkflowQuery1": &WorkflowQuery,
	}
	WorkflowQueryResult = apiv1.WorkflowQueryResult{
		ResultType:   QueryResultType,
		Answer:       &Payload1,
		ErrorMessage: ErrorMessage,
	}
	WorkflowQueryResultMap = map[string]*apiv1.WorkflowQueryResult{
		"WorkflowQuery1": &WorkflowQueryResult,
	}
	QueryRejected = apiv1.QueryRejected{
		CloseStatus: WorkflowExecutionCloseStatus,
	}
	WorkflowExecutionConfiguration = apiv1.WorkflowExecutionConfiguration{
		TaskList:                     &TaskList,
		ExecutionStartToCloseTimeout: Duration1,
		TaskStartToCloseTimeout:      Duration2,
	}
	PendingActivityInfo = apiv1.PendingActivityInfo{
		ActivityId:         ActivityID,
		ActivityType:       &ActivityType,
		State:              PendingActivityState,
		HeartbeatDetails:   &Payload1,
		LastHeartbeatTime:  Timestamp1,
		LastStartedTime:    Timestamp2,
		Attempt:            Attempt,
		MaximumAttempts:    3,
		ScheduledTime:      Timestamp3,
		ExpirationTime:     Timestamp4,
		LastFailure:        &Failure,
		LastWorkerIdentity: Identity,
	}
	PendingActivityInfoArray = []*apiv1.PendingActivityInfo{
		&PendingActivityInfo,
	}
	PendingChildExecutionInfo = apiv1.PendingChildExecutionInfo{
		WorkflowExecution: &WorkflowExecution,
		WorkflowTypeName:  WorkflowTypeName,
		InitiatedId:       EventID1,
		ParentClosePolicy: ParentClosePolicy,
	}
	PendingChildExecutionInfoArray = []*apiv1.PendingChildExecutionInfo{
		&PendingChildExecutionInfo,
	}
	PendingDecisionInfo = apiv1.PendingDecisionInfo{
		State:                 PendingDecisionState,
		ScheduledTime:         Timestamp1,
		StartedTime:           Timestamp2,
		Attempt:               Attempt,
		OriginalScheduledTime: Timestamp3,
	}
)
