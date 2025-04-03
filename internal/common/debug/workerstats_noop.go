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
	// pollerTrackerNoopImpl implements the PollerTracker interface
	pollerTrackerNoopImpl struct{}
	// stopperNoopImpl implements the Stopper interface
	stopperNoopImpl struct{}
	// activityTrackerNoopImpl implements the ActivityTracker interface
	activityTrackerNoopImpl struct{}
)

func (lc *pollerTrackerNoopImpl) Start() Stopper { return &stopperNoopImpl{} }
func (lc *pollerTrackerNoopImpl) Stats() int32   { return 0 }
func (r *stopperNoopImpl) Stop()                 {}

// NewNoopPollerTracker creates a new PollerTracker instance
func NewNoopPollerTracker() PollerTracker { return &pollerTrackerNoopImpl{} }

func (at *activityTrackerNoopImpl) Start(info ActivityInfo) Stopper { return &stopperNoopImpl{} }
func (at *activityTrackerNoopImpl) Stats() Activities               { return nil }

// NewNoopActivityTracker creates a new PollerTracker instance
func NewNoopActivityTracker() ActivityTracker { return &activityTrackerNoopImpl{} }

var _ PollerTracker = &pollerTrackerNoopImpl{}
var _ Stopper = &stopperNoopImpl{}
var _ ActivityTracker = &activityTrackerNoopImpl{}
