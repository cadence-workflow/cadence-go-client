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
	"errors"
	"sync"
	"time"

	"github.com/pborman/uuid"
	"go.uber.org/zap"
)

func init() {
	RegisterActivityWithOptions(sessionCreationActivity, RegisterActivityOptions{
		Name: sessionCreationActivityName,
	})
	RegisterActivityWithOptions(sessionCompletionActivity, RegisterActivityOptions{
		Name: sessionCompletionActivityName,
	})
}

type (
	// SessionInfo contains information of a created session. For now, exported
	// fields are SessionID and HostName.
	// SessionID is a uuid generated when CreateSession() or RecreateSession()
	// is called and can be used to uniquely identify a session.
	// HostName specifies which host is executing the session
	SessionInfo struct {
		SessionID    string
		HostName     string
		resourceID   string // hide from user for now
		tasklist     string // resource specific tasklist
		sessionState sessionState
		cancelFunc   CancelFunc
	}

	// SessionOptions specifies metadata for a session.
	// ExecutionTimeout: required, no default
	//     Specifies the maximum amount of time the session can run
	// CreationTimeout: required, no default
	//     Specfifies how long session creation can take before returning an error
	SessionOptions struct {
		ExecutionTimeout time.Duration
		CreationTimeout  time.Duration
	}

	// RecreateSessionParams contains information needed to recreate a session on the same worker.
	// Use SessionInfo.GetRecreateParameter() and pass the returned value to RecreateSession().
	RecreateSessionParams struct {
		tasklist string
	}

	sessionState int

	sessionTokenBucket struct {
		*sync.Cond
		availableToken int
	}

	sessionEnvironment interface {
		CreateSession(ctx context.Context, sessionID string) (<-chan struct{}, error)
		CompleteSession(sessionID string)
		AddSessionToken()
		SignalCreationResponse(ctx context.Context, sessionID string) error
		GetResourceSpecificTasklist() string
		GetTokenBucket() *sessionTokenBucket
	}

	sessionEnvironmentImpl struct {
		*sync.Mutex
		doneChanMap              map[string]chan struct{}
		resourceID               string
		resourceSpecificTasklist string
		sessionTokenBucket       *sessionTokenBucket
	}

	sessionCreationResponse struct {
		Tasklist   string
		HostName   string
		ResourceID string
	}
)

// Session State enum
const (
	sessionStateOpen sessionState = iota
	sessionStateFailed
	sessionStateClosed
)

const (
	sessionInfoContextKey        contextKey = "sessionInfo"
	sessionEnvironmentContextKey contextKey = "sessionEnvironment"

	sessionCreationActivityName   string = "internalSessionCreationActivity"
	sessionCompletionActivityName string = "internalSessionCompletionActivity"

	errTooManySessionsMsg string = "too many outstanding sessions"

	sessionHeartBeatTimeout time.Duration = time.Second * 20
)

var (
	// ErrSessionFailed is the error returned when user tries to execute an activity but the
	// session it belongs to has already failed
	ErrSessionFailed            = errors.New("session has failed")
	errFoundExistingOpenSession = errors.New("found exisiting open session in the context")
)

// Note: Worker should be configured to process session. To do this, set the following
// fields in WorkerOptions:
//     EnableSessionWorker: true
//     SessionResourceID: The identifier of the resource consumed by sessions.
//         It's the user's responsibility to ensure there's only one worker using this resourceID.
//         This option is not available for now as automatic session reestablishing is not implemented.
//     MaxConCurrentSessionExecutionSize: the maximum number of concurrently sessions the resource
//         support. By default, 1000 is used.

// CreateSession creates a session and returns a new context which contains information
// of the created session. The session will be created on the tasklist user specified in
// ActivityOptions. If none is specified, the default one will be used.
//
// CreationSession will fail in the following situations:
//     1. The context passed in already contains a session which is still open
//        (not closed and failed).
//     2. All the workers are busy (number of sessions currently running on all the workers have reached
//        MaxConCurrentSessionExecutionSize, which is specified when starting the workers) and session
//        cannot be created within a specified timeout.
//
// If an activity is executed using the returned context, it's regarded as part of the
// session. All activities within the same session will be executed by the same worker.
// User still needs to handle the error returned when executing an activity. Session will
// not be marked as failed if an activity within it returns an error. Only when the worker
// executing the session is down, that session will be marked as failed. Executing an activity
// within a failed session will return ErrSessionFailed immediately without scheduling that activity.
//
// If user wants to end a session since activity returns some error, use CompleteSession API below.
// New session can be created if necessary to retry the whole session.
//
// Example:
//    sessionCtx, err := CreateSession(ctx)
//    if err != nil {
//		    // Creation failed. Wrong ctx or too many outstanding sessions.
//    }
//    defer CompleteSession(sessionCtx)
//    err = ExecuteActivity(sessionCtx, someActivityFunc, activityInput).Get(sessionCtx, nil)
//    if err == ErrSessionFailed {
//        // Session has failed
//    } else {
//        // Handle activity error
//    }
//    ... // execute more activities using sessionCtx
func CreateSession(ctx Context, sessionOptions *SessionOptions) (Context, error) {
	options := getActivityOptions(ctx)
	baseTasklist := options.TaskListName
	if baseTasklist == "" {
		baseTasklist = options.OriginalTaskListName
	}
	return createSession(ctx, getCreationTasklist(baseTasklist), sessionOptions, true)
}

// RecreateSession recreate a session based on the sessionInfo passed in. Activities executed within
// the recreated session will be executed by the same worker as the previous session. RecreateSession()
// returns an error under the same situation as CreateSession() and has the same usage as CreateSession().
// It will not check the state of the session described by the sessionInfo passed in, so user can recreate
// a session based on a failed or completed session.
//
// The main usage of RecreateSession is for long sessions that are splited into multiple runs. At the end of
// one run, complete the current session, get recreateSessionParams from sessionInfo and pass the parameter to
// next run. In the new run, the session can be recreated using the parameter.
func RecreateSession(ctx Context, params *RecreateSessionParams, sessionOptions *SessionOptions) (Context, error) {
	return createSession(ctx, params.tasklist, sessionOptions, false)
}

// CompleteSession completes a session. It releases worker resources, so other sessions can be created.
// CompleteSession won't do anything if the context passed in doesn't contain any session information or the
// session has already completed or failed.
//
// After a session has completed, user can continue to use the context, but the activities will be scheduled
// on the normal taskList (as user specified in ActivityOptions) and may be picked up by another worker since
// it's not in a session.
func CompleteSession(ctx Context) {
	sessionInfo := getSessionInfo(ctx)
	if sessionInfo == nil || sessionInfo.sessionState != sessionStateOpen {
		return
	}

	retryPolicy := &RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 1.1,
		MaximumInterval:    time.Second * 10,
		MaximumAttempts:    5,
	}
	ao := ActivityOptions{
		ScheduleToStartTimeout: time.Second * 10,
		StartToCloseTimeout:    time.Second * 10,
		RetryPolicy:            retryPolicy,
	}

	// first cancel the creation activity context
	sessionInfo.cancelFunc()

	// then execute then completion activity
	completionCtx := WithActivityOptions(ctx, ao)
	// the tasklist will be overrided to use the one stored in sessionInfo
	ExecuteActivity(completionCtx, sessionCompletionActivityName, sessionInfo.SessionID)
	sessionInfo.sessionState = sessionStateClosed
	getWorkflowEnvironment(ctx).RemoveSession(sessionInfo.SessionID)
	GetLogger(ctx).Debug("Completed session", zap.String("sessionID", sessionInfo.SessionID))
}

// GetSessionInfo returns the sessionInfo stored in the context. If there are multiple sessions in the context,
// (for example, the same context is used to create, complete, create another session. Then user found that the
// session has failed, and created a new one on it), the most recent sessionInfo will be returned.
//
// This API will return nil if there's no sessionInfo in the context.
func GetSessionInfo(ctx Context) *SessionInfo {
	info := getSessionInfo(ctx)
	if info == nil {
		GetLogger(ctx).Warn("Context contains no session information")
	}
	return info
}

// GetRecreateParams returns the parameters needed to recreate a session. The returned value should be passed to
// RecreateSession() API.
func (s *SessionInfo) GetRecreateParams() *RecreateSessionParams {
	return &RecreateSessionParams{
		tasklist: s.tasklist,
	}
}

func getSessionInfo(ctx Context) *SessionInfo {
	info := ctx.Value(sessionInfoContextKey)
	if info == nil {
		return nil
	}
	return info.(*SessionInfo)
}

func setSessionInfo(ctx Context, sessionInfo *SessionInfo) Context {
	return WithValue(ctx, sessionInfoContextKey, sessionInfo)
}

func createSession(ctx Context, creationTasklist string, options *SessionOptions, retryable bool) (Context, error) {
	logger := GetLogger(ctx)
	logger.Debug("Start creating session")
	if prevSessionInfo := getSessionInfo(ctx); prevSessionInfo != nil && prevSessionInfo.sessionState == sessionStateOpen {
		return nil, errFoundExistingOpenSession
	}
	sessionID, err := generateSessionID(ctx)
	if err != nil {
		return nil, err
	}

	tasklistChan := GetSignalChannel(ctx, sessionID) // use sessionID as channel name
	// Retry is only needed when creating new session and the error returned is NewCustomError(errTooManySessionsMsg)
	retryPolicy := &RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 1.1,
		MaximumInterval:    time.Second * 10,
		ExpirationInterval: options.CreationTimeout,
		NonRetriableErrorReasons: []string{
			"cadenceInternal:Panic",
			"cadenceInternal:Generic",
			"cadenceInternal:Timeout START_TO_CLOSE",
			"cadenceInternal:Timeout HEARTBEAT",
		},
	}

	ao := ActivityOptions{
		TaskList:               creationTasklist,
		ScheduleToStartTimeout: options.CreationTimeout,
		StartToCloseTimeout:    options.ExecutionTimeout,
		HeartbeatTimeout:       sessionHeartBeatTimeout,
	}
	if retryable {
		ao.RetryPolicy = retryPolicy
	}
	creationCtx, cancel := WithCancel(WithActivityOptions(ctx, ao))
	creationFuture := ExecuteActivity(creationCtx, sessionCreationActivityName, sessionID)

	var creationErr error
	var creationResponse sessionCreationResponse
	s := NewSelector(ctx)
	s.AddReceive(tasklistChan, func(c Channel, more bool) {
		c.Receive(ctx, &creationResponse)
	})
	s.AddFuture(creationFuture, func(f Future) {
		// activity stoped before signal is received, must be creation timeout.
		creationErr = f.Get(ctx, nil)
		GetLogger(ctx).Debug("Failed to create session", zap.String("sessionID", sessionID), zap.Error(creationErr))
	})
	s.Select(ctx)

	if creationErr != nil {
		return nil, creationErr
	}

	sessionInfo := &SessionInfo{
		SessionID:    sessionID,
		tasklist:     creationResponse.Tasklist,
		resourceID:   creationResponse.ResourceID,
		HostName:     creationResponse.HostName,
		sessionState: sessionStateOpen,
		cancelFunc:   cancel,
	}

	Go(ctx, func(ctx Context) {
		if err = creationFuture.Get(ctx, nil); err != nil {
			getWorkflowEnvironment(ctx).RemoveSession(sessionID)
			GetLogger(ctx).Debug("Session failed", zap.String("sessionID", sessionID), zap.Error(err))
			sessionInfo.sessionState = sessionStateFailed
		}
	})

	logger.Debug("Created session", zap.String("sessionID", sessionID))
	getWorkflowEnvironment(ctx).AddSession(sessionInfo)
	return setSessionInfo(ctx, sessionInfo), nil
}

func generateSessionID(ctx Context) (string, error) {
	var sessionID string
	err := SideEffect(ctx, func(ctx Context) interface{} {
		return uuid.New()
	}).Get(&sessionID)
	return sessionID, err
}

func getCreationTasklist(base string) string {
	return base + "__internal_session_creation"
}

func getResourceSpecificTasklist(resourceID string) string {
	return resourceID + "@" + getHostName()
}

func sessionCreationActivity(ctx context.Context, sessionID string) error {
	sessionEnv, ok := ctx.Value(sessionEnvironmentContextKey).(sessionEnvironment)
	if !ok {
		panic("no session environment in context")
	}

	doneCh, err := sessionEnv.CreateSession(ctx, sessionID)
	if err != nil {
		return err
	}

	defer sessionEnv.AddSessionToken()

	if err := sessionEnv.SignalCreationResponse(ctx, sessionID); err != nil {
		return err
	}

	activityEnv := getActivityEnv(ctx)
	ticker := time.NewTicker(activityEnv.heartbeatTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := activityEnv.serviceInvoker.Heartbeat([]byte{})
			if err != nil {
				sessionEnv.CompleteSession(sessionID)
				return err
			}
		case <-doneCh:
			return nil
		}
	}
}

func sessionCompletionActivity(ctx context.Context, sessionID string) error {
	sessionEnv, ok := ctx.Value(sessionEnvironmentContextKey).(sessionEnvironment)
	if !ok {
		panic("no session environment in context")
	}
	sessionEnv.CompleteSession(sessionID)
	return nil
}

func isSessionCreationActivity(activity interface{}) bool {
	activityName, ok := activity.(string)
	return ok && activityName == sessionCreationActivityName
}

func newSessionTokenBucket(concurrentSessionExecutionSize int) *sessionTokenBucket {
	return &sessionTokenBucket{
		Cond:           sync.NewCond(&sync.Mutex{}),
		availableToken: concurrentSessionExecutionSize,
	}
}

func (t *sessionTokenBucket) waitForAvailableToken() {
	t.L.Lock()
	defer t.L.Unlock()
	for t.availableToken == 0 {
		t.Wait()
	}
}

func (t *sessionTokenBucket) addToken() {
	t.L.Lock()
	t.availableToken++
	t.L.Unlock()
	t.Signal()
}

func (t *sessionTokenBucket) getToken() bool {
	t.L.Lock()
	defer t.L.Unlock()
	if t.availableToken == 0 {
		return false
	}
	t.availableToken--
	return true
}

func newSessionEnvironment(resourceID string, concurrentSessionExecutionSize int) sessionEnvironment {
	return &sessionEnvironmentImpl{
		Mutex:                    &sync.Mutex{},
		doneChanMap:              make(map[string]chan struct{}),
		resourceID:               resourceID,
		resourceSpecificTasklist: getResourceSpecificTasklist(resourceID),
		sessionTokenBucket:       newSessionTokenBucket(concurrentSessionExecutionSize),
	}
}

func (env *sessionEnvironmentImpl) CreateSession(ctx context.Context, sessionID string) (<-chan struct{}, error) {
	if !env.sessionTokenBucket.getToken() {
		return nil, NewCustomError(errTooManySessionsMsg)
	}

	env.Lock()
	defer env.Unlock()
	doneCh := make(chan struct{})
	env.doneChanMap[sessionID] = doneCh
	return doneCh, nil
}

func (env *sessionEnvironmentImpl) AddSessionToken() {
	env.sessionTokenBucket.addToken()
}

func (env *sessionEnvironmentImpl) SignalCreationResponse(ctx context.Context, sessionID string) error {
	activityEnv := getActivityEnv(ctx)
	client := activityEnv.serviceInvoker.GetClient(activityEnv.workflowDomain, &ClientOptions{})
	return client.SignalWorkflow(ctx, activityEnv.workflowExecution.ID, activityEnv.workflowExecution.RunID,
		sessionID, env.getCreationResponse())
}

func (env *sessionEnvironmentImpl) getCreationResponse() *sessionCreationResponse {
	return &sessionCreationResponse{
		Tasklist:   env.resourceSpecificTasklist,
		ResourceID: env.resourceID,
		HostName:   getHostName(),
	}
}

func (env *sessionEnvironmentImpl) CompleteSession(sessionID string) {
	env.Lock()
	defer env.Unlock()

	if doneChan, ok := env.doneChanMap[sessionID]; ok {
		delete(env.doneChanMap, sessionID)
		close(doneChan)
	}
}

func (env *sessionEnvironmentImpl) GetResourceSpecificTasklist() string {
	return env.resourceSpecificTasklist
}

func (env *sessionEnvironmentImpl) GetTokenBucket() *sessionTokenBucket {
	return env.sessionTokenBucket
}