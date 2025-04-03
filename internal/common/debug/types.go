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

type (
	// Stopper is an interface for tracking stop events in an ongoing process or goroutine.
	// Implementations should ensure not to clean up any resources opened by worker
	// Deprecated: in development and very likely to change
	Stopper interface {
		// Stop signals and track a stop event
		Stop()
	}

	// PollerTracker is an interface to track running pollers on a worker
	// Deprecated: in development and very likely to change
	PollerTracker interface {
		// Start collects information on poller start up.
		// consumers should provide a concurrency-safe implementation.
		Start() Stopper
		// Stats return the number or running pollers
		Stats() int32
	}

	// WorkerStats provides a set of methods that can be used to collect
	// stats on the Worker for debugging purposes.
	// Deprecated: in development and very likely to change
	WorkerStats struct {
		PollerTracker   PollerTracker
		ActivityTracker ActivityTracker
	}

	// ActivityInfo contains details on the executing activity
	// Deprecated: in development and very likely to change
	ActivityInfo struct {
		TaskList     string
		ActivityType string
	}

	// ActivityTracker is a worker option to track executing activities on a worker
	// Deprecated: in development and very likely to change
	ActivityTracker interface {
		// Start records activity execution
		Start(info ActivityInfo) Stopper
		// Stats returns a list of executing activity info
		Stats() Activities
	}

	// Activities is a list of executing activities on the worker
	// Deprecated: in development and very likely to change
	Activities []struct {
		Info  ActivityInfo
		Count int64
	}

	// Debugger exposes stats collected on a running Worker
	// Deprecated: in development and very likely to change
	Debugger interface {
		GetWorkerStats() WorkerStats
	}
)
