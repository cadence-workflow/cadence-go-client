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
)

type (
	// pollerTrackerImpl implements the PollerTracker interface
	pollerTrackerImpl struct {
		sync.RWMutex
		count map[string]int64
	}

	// stopperImpl implements the Stopper interface
	stopperImpl struct {
		sync.Once
		workerType    string
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

func (p *pollerTrackerImpl) Start(workerType string) Stopper {
	p.Lock()
	defer p.Unlock()
	p.count[workerType]++
	return &stopperImpl{pollerTracker: p, workerType: workerType}
}

func (p *pollerTrackerImpl) Stats() Pollers {
	var pollers Pollers
	p.RLock()
	defer p.RUnlock()
	for pollerType, count := range p.count {
		if count > 0 {
			pollers = append(pollers, struct {
				Type  string
				Count int64
			}{Type: pollerType, Count: count})
		}
	}
	sort.Slice(pollers, func(i, j int) bool {
		return pollers[i].Type < pollers[j].Type
	})
	return pollers
}

func (s *stopperImpl) Stop() {
	s.Do(func() {
		s.pollerTracker.Lock()
		defer s.pollerTracker.Unlock()
		s.pollerTracker.count[s.workerType]--
		if s.pollerTracker.count[s.workerType] == 0 {
			delete(s.pollerTracker.count, s.workerType)
		}
	})
}

func Example() {
	var pollerTracker PollerTracker
	pollerTracker = &pollerTrackerImpl{count: make(map[string]int64)}

	// Initially, poller count should be 0
	jsonPollers, _ := json.MarshalIndent(pollerTracker.Stats(), "", "  ")
	fmt.Println(string(jsonPollers))

	// Start pollers and verify that the count increments
	stopper1 := pollerTracker.Start("ActivityWorker")
	jsonPollers, _ = json.MarshalIndent(pollerTracker.Stats(), "", "  ")
	fmt.Println(string(jsonPollers))
	stopper2 := pollerTracker.Start("ActivityWorker")
	jsonPollers, _ = json.MarshalIndent(pollerTracker.Stats(), "", "  ")
	fmt.Println(string(jsonPollers))
	// Start another poller type and verify that the count increments
	stopper3 := pollerTracker.Start("WorkflowWorker")
	jsonPollers, _ = json.MarshalIndent(pollerTracker.Stats(), "", "  ")
	fmt.Println(string(jsonPollers))
	// Stop the pollers and verify the counter
	stopper1.Stop()
	stopper2.Stop()
	stopper3.Stop()
	jsonPollers, _ = json.MarshalIndent(pollerTracker.Stats(), "", "  ")
	fmt.Println(string(jsonPollers))

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
	// null
	// [
	//   {
	//     "Type": "ActivityWorker",
	//     "Count": 1
	//   }
	// ]
	// [
	//   {
	//     "Type": "ActivityWorker",
	//     "Count": 2
	//   }
	// ]
	// [
	//   {
	//     "Type": "ActivityWorker",
	//     "Count": 2
	//   },
	//   {
	//     "Type": "WorkflowWorker",
	//     "Count": 1
	//   }
	// ]
	// null
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
