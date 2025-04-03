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
	"sync"
	"testing"
	"time"

	"go.uber.org/cadence/internal/common/testlogger"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/util"
)

func TestAutoHearbeater_Run(t *testing.T) {
	var (
		activityType      = "test-activity"
		workflowExecution = &shared.WorkflowExecution{
			WorkflowId: common.StringPtr("test-workflow"),
			RunId:      common.StringPtr("test-run"),
		}
	)

	// Run is a sync function that is spawned in a goroutine normally.
	// So instead of asserting the results we can verify that the function exits.

	t.Run("worker stop channel", func(t *testing.T) {
		stopCh := make(chan struct{})
		invoker := &MockServiceInvoker{}
		logger := testlogger.NewZap(t)
		clock := clockwork.NewFakeClock()
		hearbeater := newHeartbeater(stopCh, invoker, logger, clock, activityType, workflowExecution)

		close(stopCh)

		hearbeater.Run(context.Background(), time.Second)
	})
	t.Run("context done", func(t *testing.T) {
		stopCh := make(chan struct{})
		invoker := &MockServiceInvoker{}
		logger := testlogger.NewZap(t)
		clock := clockwork.NewFakeClock()
		hearbeater := newHeartbeater(stopCh, invoker, logger, clock, activityType, workflowExecution)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		hearbeater.Run(ctx, time.Second)
	})
	t.Run("heartbeat success", func(t *testing.T) {
		stopCh := make(chan struct{})
		invoker := &MockServiceInvoker{}
		invoker.EXPECT().BackgroundHeartbeat().Return(nil).Once()
		logger := testlogger.NewZap(t)
		clock := clockwork.NewFakeClock()
		hearbeater := newHeartbeater(stopCh, invoker, logger, clock, activityType, workflowExecution)

		ctx, cancel := context.WithCancel(context.Background())

		wg := &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			hearbeater.Run(ctx, time.Second)
			wg.Done()
		}()
		clock.BlockUntil(1)
		clock.Advance(time.Second / 2)
		cancel()

		require.True(t, util.AwaitWaitGroup(wg, time.Second))
	})
	t.Run("heartbeat fail", func(t *testing.T) {
		stopCh := make(chan struct{})
		invoker := &MockServiceInvoker{}
		invoker.EXPECT().BackgroundHeartbeat().Return(assert.AnError).Once()
		logger := testlogger.NewZap(t)
		clock := clockwork.NewFakeClock()
		hearbeater := newHeartbeater(stopCh, invoker, logger, clock, activityType, workflowExecution)

		ctx, cancel := context.WithCancel(context.Background())

		wg := &sync.WaitGroup{}
		wg.Add(1)

		go func() {
			hearbeater.Run(ctx, time.Second)
			wg.Done()
		}()
		clock.BlockUntil(1)
		clock.Advance(time.Second / 2)
		cancel()

		require.True(t, util.AwaitWaitGroup(wg, time.Second))
	})
}
