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
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"go.uber.org/atomic"
)

type (
	// pollerTrackerImpl implements the PollerTracker interface
	pollerTrackerImpl struct {
		pollerCount atomic.Int32
	}

	// stopperImpl implements the Stopper interface
	stopperImpl struct {
		pollerTracker *pollerTrackerImpl
	}

	// activityTrackerImpl implements the ActivityTracker interface
	activityTrackerImpl struct {
		sync.RWMutex
		activityCount map[ActivityInfo]int64
	}

	// activityStopperImpl implements the Stopper interface
	activityStopperImpl struct {
		sync.Once
		info    ActivityInfo
		tracker *activityTrackerImpl
	}
)

var _ ActivityTracker = &activityTrackerImpl{}
var _ Stopper = &activityStopperImpl{}

func (ati *activityTrackerImpl) Start(info ActivityInfo) Stopper {
	ati.Lock()
	defer ati.Unlock()
	ati.activityCount[info]++
	return &activityStopperImpl{info: info, tracker: ati}
}

func (ati *activityTrackerImpl) Stats() Activities {
	var activities Activities
	ati.RLock()
	defer ati.RUnlock()
	for a, count := range ati.activityCount {
		if count > 0 {
			activities = append(activities, struct {
				Info  ActivityInfo
				Count int64
			}{Info: a, Count: count})
		}
	}
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Info.ActivityType < activities[j].Info.ActivityType
	})
	return activities
}

func (asi *activityStopperImpl) Stop() {
	asi.Do(func() {
		asi.tracker.Lock()
		defer asi.tracker.Unlock()
		asi.tracker.activityCount[asi.info]--
		if asi.tracker.activityCount[asi.info] == 0 {
			delete(asi.tracker.activityCount, asi.info)
		}
	})
}

func (p *pollerTrackerImpl) Start() Stopper {
	p.pollerCount.Inc()
	return &stopperImpl{
		pollerTracker: p,
	}
}

func (p *pollerTrackerImpl) Stats() int32 {
	return p.pollerCount.Load()
}

func (s *stopperImpl) Stop() {
	s.pollerTracker.pollerCount.Dec()
}

func Example() {
	var pollerTracker PollerTracker
	pollerTracker = &pollerTrackerImpl{}

	// Initially, poller count should be 0
	fmt.Println(fmt.Sprintf("poller stats: %d", pollerTracker.Stats()))

	// Start a poller and verify that the count increments
	stopper1 := pollerTracker.Start()
	fmt.Println(fmt.Sprintf("poller stats: %d", pollerTracker.Stats()))

	// Start another poller and verify that the count increments again
	stopper2 := pollerTracker.Start()
	fmt.Println(fmt.Sprintf("poller stats: %d", pollerTracker.Stats()))

	// Stop the pollers and verify the counter
	stopper1.Stop()
	stopper2.Stop()
	fmt.Println(fmt.Sprintf("poller stats: %d", pollerTracker.Stats()))

	var activityTracker ActivityTracker
	activityTracker = &activityTrackerImpl{activityCount: make(map[ActivityInfo]int64)}

	info1 := ActivityInfo{
		TaskList:     "task-list",
		ActivityType: "activity1",
	}

	info2 := ActivityInfo{
		TaskList:     "task-list",
		ActivityType: "activity2",
	}

	stopper1 = activityTracker.Start(info1)
	stopper2 = activityTracker.Start(info2)
	jsonActivities, _ := json.MarshalIndent(activityTracker.Stats(), "", "  ")
	fmt.Println(string(jsonActivities))

	stopper1.Stop()
	stopper1.Stop()
	jsonActivities, _ = json.MarshalIndent(activityTracker.Stats(), "", "  ")

	fmt.Println(string(jsonActivities))
	stopper2.Stop()

	jsonActivities, _ = json.MarshalIndent(activityTracker.Stats(), "", "  ")
	fmt.Println(string(jsonActivities))

	// Output:
	// poller stats: 0
	// poller stats: 1
	// poller stats: 2
	// poller stats: 0
	// [
	//   {
	//     "Info": {
	//       "TaskList": "task-list",
	//       "ActivityType": "activity1"
	//     },
	//     "Count": 1
	//   },
	//   {
	//     "Info": {
	//       "TaskList": "task-list",
	//       "ActivityType": "activity2"
	//     },
	//     "Count": 1
	//   }
	// ]
	// [
	//   {
	//     "Info": {
	//       "TaskList": "task-list",
	//       "ActivityType": "activity2"
	//     },
	//     "Count": 1
	//   }
	// ]
	// null
}
