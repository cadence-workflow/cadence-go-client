// Copyright (c) 2017-2021 Uber Technologies Inc.
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
	"time"

	"github.com/jonboulle/clockwork"
	"go.uber.org/zap"

	"go.uber.org/cadence/.gen/go/shared"
)

func newHeartbeater(
	stopChannel <-chan struct{},
	invoker ServiceInvoker,
	logger *zap.Logger,
	clock clockwork.Clock,
	activityType string,
	execution *shared.WorkflowExecution,
) *autoHeartbeater {
	return &autoHeartbeater{
		stopCh:       stopChannel,
		invoker:      invoker,
		logger:       logger,
		clock:        clock,
		activityType: activityType,
		execution:    execution,
	}
}

type autoHeartbeater struct {
	stopCh       <-chan struct{}
	invoker      ServiceInvoker
	logger       *zap.Logger
	clock        clockwork.Clock
	activityType string
	execution    *shared.WorkflowExecution
}

func (a *autoHeartbeater) Run(ctx context.Context, timeout time.Duration) {
	autoHbInterval := timeout / 2
	ticker := a.clock.NewTicker(autoHbInterval)
	defer ticker.Stop()
	for {
		select {
		case <-a.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.Chan():
			hbErr := a.invoker.BackgroundHeartbeat()
			if hbErr != nil && !IsCanceledError(hbErr) {
				a.logger.Error("Activity auto heartbeat error.",
					zap.String(tagWorkflowID, a.execution.GetWorkflowId()),
					zap.String(tagRunID, a.execution.GetRunId()),
					zap.String(tagActivityType, a.activityType),
					zap.Error(hbErr),
				)
			}
		}
	}
}
