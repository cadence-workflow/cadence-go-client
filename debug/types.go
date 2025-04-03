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

package debug

import (
	internal "go.uber.org/cadence/internal/common/debug"
)

type (
	// PollerTracker is an interface to track running pollers on a worker
	// Deprecated: in development and very likely to change
	PollerTracker = internal.PollerTracker

	// Stopper is an interface for tracking stop events in an ongoing process or goroutine.
	// Implementations should ensure that Stop() is used to signal and track the stop event
	// and not to clean up any resources opened by worker
	// Deprecated: in development and very likely to change
	Stopper = internal.Stopper

	// WorkerStats provides a set of methods that can be used to collect
	// stats on the Worker for debugging purposes.
	// Deprecated: in development and very likely to change
	WorkerStats = internal.WorkerStats

	// ActivityTracker is a worker option to track executing activities on a worker
	// Deprecated: in development and very likely to change
	ActivityTracker = internal.ActivityTracker

	// ActivityInfo contains details on the executing activity
	// Deprecated: in development and very likely to change
	ActivityInfo = internal.ActivityInfo

	// Activities is a list of executing activities on the worker
	// Deprecated: in development and very likely to change
	Activities = internal.Activities

	// Debugger exposes stats collected on a running Worker
	// Deprecated: in development and very likely to change
	Debugger = internal.Debugger
)
