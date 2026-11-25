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
	"time"
)

// defaultPollerScalerCooldownInSeconds
const (
	defaultPollerAutoScalerCooldown           = 10 * time.Second
	defaultMinPollerSize                      = 2
	defaultMaxPollerSize                      = 20
	defaultInitPollerSize                     = 2
	defaultPollerAutoScalerWaitTimeUpperBound = 256 * time.Millisecond
	defaultPollerAutoScalerWaitTimeLowerBound = 16 * time.Millisecond
)

type (
	AutoScalerOptions struct {
		// Optional: Enable the auto scaler.
		// default: false
		Enabled bool

		// Optional: The cooldown period after a scale up or down.
		// default: 10 seconds
		Cooldown time.Duration

		// Optional: The minimum number of pollers to start with.
		// default: 2
		PollerMinCount int

		// Optional: The maximum number of pollers to start with.
		// default: 20
		PollerMaxCount int

		// Optional: The initial number of pollers to start with.
		// default: 2
		PollerInitCount int

		// Optional: The upper bound of poller wait time for poller autoscaler to scale down.
		// default: 256ms
		// NOTE: This is normally not needed to be set by user.
		PollerWaitTimeUpperBound time.Duration

		// Optional: The lower bound of poller wait time for poller autoscaler to scale up.
		// default: 16ms
		// NOTE: This is normally not needed to be set by user.
		PollerWaitTimeLowerBound time.Duration
	}
)
