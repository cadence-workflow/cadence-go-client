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
	"errors"
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"

	"go.uber.org/cadence/internal/common/autoscaler"
	"go.uber.org/cadence/internal/worker"
)

// defaultPollerScalerCooldownInSeconds
const (
	defaultPollerAutoScalerCooldown          = time.Minute
	defaultPollerAutoScalerTargetUtilization = 0.6
	defaultMinConcurrentActivityPollerSize   = 1
	defaultMinConcurrentDecisionPollerSize   = 2
)

var (
	_ autoscaler.AutoScaler = (*pollerAutoScaler)(nil)
	_ autoscaler.Estimator  = (*pollerUsageEstimator)(nil)
)

type (
	pollerAutoScaler struct {
		pollerUsageEstimator

		isDryRun     bool
		cooldownTime time.Duration
		logger       *zap.Logger
		permit       worker.Permit
		ctx          context.Context
		cancel       context.CancelFunc
		wg           *sync.WaitGroup // graceful stop
		recommender  autoscaler.Recommender
		onAutoScale  []func() // hook functions that run post autoscale
	}

	pollerUsageEstimator struct {
		// This single atomic variable stores two variables:
		// left 32 bits is noTaskCounts, right 32 bits is taskCounts.
		// This avoids unnecessary usage of CompareAndSwap
		atomicBits *atomic.Uint64
	}

	pollerAutoScalerOptions struct {
		Enabled           bool
		InitCount         int
		MinCount          int
		MaxCount          int
		Cooldown          time.Duration
		DryRun            bool
		TargetUtilization float64
	}
)

func newPollerScaler(
	options pollerAutoScalerOptions,
	logger *zap.Logger,
	permit worker.Permit,
	hooks ...func()) *pollerAutoScaler {
	if !options.Enabled {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &pollerAutoScaler{
		isDryRun:             options.DryRun,
		cooldownTime:         options.Cooldown,
		logger:               logger,
		permit:               permit,
		wg:                   &sync.WaitGroup{},
		ctx:                  ctx,
		cancel:               cancel,
		pollerUsageEstimator: pollerUsageEstimator{atomicBits: atomic.NewUint64(0)},
		recommender: autoscaler.NewLinearRecommender(
			autoscaler.ResourceUnit(options.MinCount),
			autoscaler.ResourceUnit(options.MaxCount),
			autoscaler.Usages{
				autoscaler.PollerUtilizationRate: autoscaler.MilliUsage(options.TargetUtilization * 1000),
			},
		),
		onAutoScale: hooks,
	}
}

// Start an auto-scaler go routine and returns a done to stop it
func (p *pollerAutoScaler) Start() {
	logger := p.logger.Sugar()
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			select {
			case <-p.ctx.Done():
				return
			case <-time.After(p.cooldownTime):
				currentResource := autoscaler.ResourceUnit(p.permit.Quota())
				currentUsages, err := p.pollerUsageEstimator.Estimate()
				if err != nil {
					logger.Warnw("poller autoscaler skip due to estimator error", "error", err)
					continue
				}
				proposedResource := p.recommender.Recommend(currentResource, currentUsages)
				logger.Debugw("poller autoscaler recommendation",
					"currentUsage", currentUsages,
					"current", uint64(currentResource),
					"recommend", uint64(proposedResource),
					"isDryRun", p.isDryRun)
				if !p.isDryRun {
					p.permit.SetQuota(int(proposedResource))
				}
				p.pollerUsageEstimator.Reset()

				// hooks
				for i := range p.onAutoScale {
					p.onAutoScale[i]()
				}
			}
		}
	}()
	return
}

// Stop stops the poller autoscaler
func (p *pollerAutoScaler) Stop() {
	p.cancel()
	p.wg.Wait()
}

// Reset metrics from the start
func (m *pollerUsageEstimator) Reset() {
	m.atomicBits.Store(0)
}

// CollectUsage counts past poll results to estimate autoscaler.Usages
func (m *pollerUsageEstimator) CollectUsage(data interface{}) error {
	isEmpty, err := isTaskEmpty(data)
	if err != nil {
		return err
	}
	if isEmpty { // no-task poll
		m.atomicBits.Add(1 << 32)
	} else {
		m.atomicBits.Add(1)
	}
	return nil
}

func isTaskEmpty(task interface{}) (bool, error) {
	switch t := task.(type) {
	case *workflowTask:
		return t == nil || t.task == nil || len(t.task.TaskToken) == 0, nil
	case *activityTask:
		return t == nil || t.task == nil || len(t.task.TaskToken) == 0, nil
	case *localActivityTask:
		return t == nil || t.workflowTask == nil, nil
	default:
		return false, errors.New("unknown task type")
	}
}

// Estimate is based on past poll counts
func (m *pollerUsageEstimator) Estimate() (autoscaler.Usages, error) {
	bits := m.atomicBits.Load()
	noTaskCounts := bits >> 32           // left 32 bits
	taskCounts := bits & ((1 << 32) - 1) // right 32 bits
	if noTaskCounts+taskCounts == 0 {
		return nil, errors.New("autoscaler.Estimator::Estimate error: not enough data")
	}

	return autoscaler.Usages{
		autoscaler.PollerUtilizationRate: autoscaler.MilliUsage(taskCounts * 1000 / (noTaskCounts + taskCounts)),
	}, nil
}
