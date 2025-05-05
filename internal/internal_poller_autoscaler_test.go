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
	"math/rand"
	"sync"
	"testing"
	"time"

	"go.uber.org/cadence/internal/common/testlogger"
	"go.uber.org/cadence/internal/worker"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"

	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common/autoscaler"
)

func Test_pollerAutoscaler(t *testing.T) {
	type args struct {
		disabled                        bool
		noTaskPoll, taskPoll, unrelated int
		initialPollerCount              int
		minPollerCount                  int
		maxPollerCount                  int
		targetMilliUsage                uint64
		cooldownTime                    time.Duration
		autoScalerEpoch                 int
		isDryRun                        bool
	}

	coolDownTime := time.Millisecond * 50

	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "dry run doesn't change anything",
			args: args{
				noTaskPoll:         100,
				taskPoll:           0,
				unrelated:          0,
				initialPollerCount: 10,
				minPollerCount:     2,
				maxPollerCount:     10,
				targetMilliUsage:   500,
				cooldownTime:       coolDownTime,
				autoScalerEpoch:    1,
				isDryRun:           true,
			},
			want: 10,
		},
		{
			name: "no utilization, scale to min",
			args: args{
				noTaskPoll:         100,
				taskPoll:           0,
				unrelated:          0,
				initialPollerCount: 10,
				minPollerCount:     1,
				maxPollerCount:     10,
				targetMilliUsage:   500,
				cooldownTime:       coolDownTime,
				autoScalerEpoch:    1,
				isDryRun:           false,
			},
			want: 1,
		},
		{
			name: "low utilization, scale down",
			args: args{
				noTaskPoll:         75,
				taskPoll:           25,
				unrelated:          0,
				initialPollerCount: 10,
				minPollerCount:     1,
				maxPollerCount:     10,
				targetMilliUsage:   500,
				cooldownTime:       coolDownTime,
				autoScalerEpoch:    1,
				isDryRun:           false,
			},
			want: 5,
		},
		{
			name: "over utilized, scale up",
			args: args{
				noTaskPoll:         0,
				taskPoll:           100,
				unrelated:          0,
				initialPollerCount: 2,
				minPollerCount:     1,
				maxPollerCount:     10,
				targetMilliUsage:   500,
				cooldownTime:       coolDownTime,
				autoScalerEpoch:    1,
				isDryRun:           false,
			},
			want: 10,
		},
		{
			name: "over utilized, scale up to max",
			args: args{
				noTaskPoll:         0,
				taskPoll:           100,
				unrelated:          0,
				initialPollerCount: 6,
				minPollerCount:     1,
				maxPollerCount:     10,
				targetMilliUsage:   500,
				cooldownTime:       coolDownTime,
				autoScalerEpoch:    1,
				isDryRun:           false,
			},
			want: 10,
		},
		{
			name: "over utilized, but wait time less than cooldown time",
			args: args{
				noTaskPoll:         0,
				taskPoll:           100,
				unrelated:          0,
				initialPollerCount: 6,
				minPollerCount:     1,
				maxPollerCount:     10,
				targetMilliUsage:   500,
				cooldownTime:       coolDownTime,
				autoScalerEpoch:    0,
				isDryRun:           false,
			},
			want: 6,
		},
		{
			name: "disabled",
			args: args{disabled: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			autoscalerEpoch := atomic.NewUint64(0)
			pollerScaler := newPollerScaler(
				pollerAutoScalerOptions{
					Enabled:           !tt.args.disabled,
					InitCount:         tt.args.initialPollerCount,
					MinCount:          tt.args.minPollerCount,
					MaxCount:          tt.args.maxPollerCount,
					Cooldown:          tt.args.cooldownTime,
					DryRun:            tt.args.isDryRun,
					TargetUtilization: float64(tt.args.targetMilliUsage) / 1000,
				},
				testlogger.NewZap(t),
				worker.NewResizablePermit(tt.args.initialPollerCount),
				// hook function that collects number of iterations
				func() {
					autoscalerEpoch.Add(1)
				})
			if tt.args.disabled {
				assert.Nil(t, pollerScaler)
				return
			}

			pollerScaler.Start()

			// simulate concurrent polling
			pollChan := generateRandomPollResults(tt.args.noTaskPoll, tt.args.taskPoll, tt.args.unrelated)
			wg := &sync.WaitGroup{}
			for i := 0; i < tt.args.maxPollerCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for pollResult := range pollChan {
						err := pollerScaler.permit.Acquire(context.Background())
						assert.NoError(t, err)
						pollerScaler.CollectUsage(pollResult)
						pollerScaler.permit.Release()
					}
				}()
			}

			assert.Eventually(t, func() bool {
				return autoscalerEpoch.Load() == uint64(tt.args.autoScalerEpoch)
			}, tt.args.cooldownTime+100*time.Millisecond, 10*time.Millisecond)
			pollerScaler.Stop()
			res := pollerScaler.permit.Quota() - pollerScaler.permit.Count()
			assert.Equal(t, tt.want, int(res))
		})
	}
}

func Test_pollerUsageEstimator(t *testing.T) {
	type args struct {
		noTaskPoll, taskPoll, unrelated int
		pollerCount                     int
	}
	tests := []struct {
		name    string
		args    args
		want    autoscaler.Usages
		wantErr bool
	}{
		{
			name: "400 no-task, 100 task, 100 unrelated",
			args: args{
				noTaskPoll:  400,
				taskPoll:    100,
				unrelated:   100,
				pollerCount: 5,
			},
			want: autoscaler.Usages{
				autoscaler.PollerUtilizationRate: 200,
			},
		},
		{
			name: "0 no-task, 0 task, 100 unrelated",
			args: args{
				noTaskPoll:  0,
				taskPoll:    0,
				unrelated:   100,
				pollerCount: 5,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimator := &pollerUsageEstimator{atomicBits: atomic.NewUint64(0)}
			pollChan := generateRandomPollResults(tt.args.noTaskPoll, tt.args.taskPoll, tt.args.unrelated)
			wg := &sync.WaitGroup{}
			for i := 0; i < tt.args.pollerCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for pollResult := range pollChan {
						estimator.CollectUsage(pollResult)
					}
				}()
			}
			wg.Wait()

			res, err := estimator.Estimate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, res)
			}
		})
	}
}

type unrelatedPolledTask struct{}

func generateRandomPollResults(noTaskPoll, taskPoll, unrelated int) <-chan interface{} {
	var result []interface{}
	for i := 0; i < noTaskPoll; i++ {
		result = append(result, &activityTask{task: &s.PollForActivityTaskResponse{}})
	}
	for i := 0; i < taskPoll; i++ {
		result = append(result, &activityTask{task: &s.PollForActivityTaskResponse{TaskToken: []byte("some value")}})
	}
	for i := 0; i < unrelated; i++ {
		result = append(result, &unrelatedPolledTask{})
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	pollChan := make(chan interface{}, len(result))
	defer close(pollChan)
	for i := range result {
		pollChan <- result[i]
	}
	return pollChan
}
