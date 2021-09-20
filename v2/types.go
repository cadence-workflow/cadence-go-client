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

package v2

import (
	"time"

	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/internal"
)

type (
	BadBinary struct {
		Checksum string
		Reason   string
		Operator string
		Created  time.Time
	}

	TaskListInfo struct {
		Pollers []PollerInfo
	}

	PollerInfo struct {
		LastAccess    time.Time
		Identity      string
		RatePerSecond *float64
	}

	DomainInfo struct {
		ID          string
		Name        string
		Status      DomainStatus
		Description string
		OwnerEmail  string
		Data        map[string]string

		WorkflowExecutionRetentionPeriod time.Duration
		HistoryArchivalStatus            ArchivalStatus
		HistoryArchivalURI               string
		VisibilityArchivalStatus         ArchivalStatus
		VisibilityArchivalURI            string
	}

	WorkflowInfo struct {
		WorkflowID       string
		RunID            string
		WorkflowType     string
		StartTime        time.Time
		CloseTime        *time.Time
		CloseStatus      *WorkflowExecutionCloseStatus
		HistoryLength    int64
		ParentWorkflowID string
		ParentRunID      string
		ParentDomainID   string
		ExecutionTime    time.Time
		Memo             map[string]encoded.Value
		SearchAttributes map[string]encoded.Value
		AutoResetPoints  []ResetPointInfo
		TaskList         string
		IsCron           bool
	}

	PendingActivityInfo struct {
		ActivityID             string
		ActivityType           string
		State                  PendingActivityState
		HeartbeatDetails       encoded.Value
		LastHeartbeatTimestamp *time.Time
		LastStartedTimestamp   *time.Time
		Attempt                int32
		MaximumAttempts        int32
		ScheduledTimestamp     *time.Time
		ExpirationTimestamp    *time.Time
		LastFailureReason      *string
		LastWorkerIdentity     *string
		LastFailureDetails     encoded.Value
	}

	PendingChildWorkflowInfo struct {
		WorkflowID        string
		RunID             string
		WorkflowType      string
		InitiatedID       int64
		ParentClosePolicy ParentClosePolicy
	}

	PendingDecisionInfo struct {
		State                      PendingDecisionState
		ScheduledTimestamp         *time.Time
		StartedTimestamp           *time.Time
		Attempt                    int64
		OriginalScheduledTimestamp *time.Time
	}

	ResetPointInfo struct {
		BinaryChecksum           string
		RunId                    string
		FirstDecisionCompletedId *int64
		CreatedTime              *time.Time
		ExpiringTime             *time.Time
		Resettable               bool
	}

	SearchAttribute struct {
		Key  string
		Type IndexedValueType
	}

	RetryPolicy = internal.RetryPolicy
)

type DomainStatus int

const (
	DomainStatusRegistered DomainStatus = iota
	DomainStatusDeprecated
	DomainStatusDeleted
)

type ArchivalStatus int

const (
	ArchivalStatusDisabled ArchivalStatus = iota
	ArchivalStatusEnabled
)

type TaskListType int

const (
	TaskListTypeDecision TaskListType = iota
	TaskListTypeActivity
)

type QueryRejectCondition int

const (
	QueryRejectConditionNotOpen QueryRejectCondition = iota
	QueryRejectConditionNotCompletedCleanly
)

type QueryConsistencyLevel int

const (
	QueryConsistencyLevelEventual QueryConsistencyLevel = iota
	QueryConsistencyLevelStrong
)

type WorkflowIDReusePolicy int

const (
	WorkflowIdReusePolicyAllowDuplicateFailedOnly WorkflowIDReusePolicy = iota
	WorkflowIdReusePolicyAllowDuplicate
	WorkflowIdReusePolicyRejectDuplicate
	WorkflowIdReusePolicyTerminateIfRunning
)

type WorkflowExecutionCloseStatus int

const (
	WorkflowExecutionCloseStatusCompleted WorkflowExecutionCloseStatus = iota
	WorkflowExecutionCloseStatusFailed
	WorkflowExecutionCloseStatusCanceled
	WorkflowExecutionCloseStatusTerminated
	WorkflowExecutionCloseStatusContinuedAsNew
	WorkflowExecutionCloseStatusTimedOut
)

type PendingActivityState int

const (
	PendingActivityStateScheduled PendingActivityState = iota
	PendingActivityStateStarted
	PendingActivityStateCancelRequested
)

type PendingDecisionState int

const (
	PendingDecisionStateScheduled PendingDecisionState = iota
	PendingDecisionStateStarted
)

type ParentClosePolicy int

const (
	ParentClosePolicyAbandon ParentClosePolicy = iota
	ParentClosePolicyRequestCancel
	ParentClosePolicyTerminate
)

type IndexedValueType int

const (
	IndexedValueTypeString IndexedValueType = iota
	IndexedValueTypeKeyword
	IndexedValueTypeInt
	IndexedValueTypeDouble
	IndexedValueTypeBool
	IndexedValueTypeDatetime
)
