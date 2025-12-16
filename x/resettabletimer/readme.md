### Resettable Timers in Cadence Workflows

#### Status

November 27, 2025

This is experimental and the API may change in future releases.

#### Background

In Cadence workflows, timers are a fundamental building block for implementing timeouts and delays. However, standard timers cannot be reset once created - you must cancel the old timer and create a new one, which can lead to complex code patterns.

The resettable timer provides a simple way to implement timeout patterns that need to restart based on external events.

#### Getting Started

Import the package:

```go
import (
    "go.uber.org/cadence/workflow"
    "go.uber.org/cadence/x/resettabletimer"
)
```

#### Basic Usage

Create a timer that can be reset:

```go
func MyWorkflow(ctx workflow.Context) error {
    // Create a timer that fires after 30 seconds
    timer := resettabletimer.New(ctx, 30*time.Second)
    
    // Wait for the timer
    err := timer.Future.Get(ctx, nil)
    if err != nil {
        return err
    }
    
    // Timer fired - handle timeout
    workflow.GetLogger(ctx).Info("Timeout occurred")
    return nil
}
```

#### Resetting the Timer

```go
func MyWorkflow(ctx workflow.Context) error {
    timer := resettabletimer.New(ctx, 30*time.Second)
    userActivityChan := workflow.GetSignalChannel(ctx, "user_activity")
    
    selector := workflow.NewSelector(ctx)
    
    // Add timer to selector
    selector.AddFuture(timer.Future, func(f workflow.Future) {
        workflow.GetLogger(ctx).Info("User inactive for 30 seconds")
    })
    
    // Add signal channel to selector
    selector.AddReceive(userActivityChan, func(c workflow.Channel, more bool) {
        var signal string
        c.Receive(ctx, &signal)
        
        // Reset the timer when activity is detected
        timer.Reset(30 * time.Second)
        workflow.GetLogger(ctx).Info("Activity detected, timer reset")
    })
    
    selector.Select(ctx)
    return nil
}
```

#### Example: Inactivity Timeout with Dynamic Duration

```go
func InactivityTimeoutWorkflow(ctx workflow.Context) error {
    // Start with 5 minute timeout
    timeout := 5 * time.Minute
    timer := resettabletimer.New(ctx, timeout)
    
    userActivityChan := workflow.GetSignalChannel(ctx, "user_activity")
    stopChan := workflow.GetSignalChannel(ctx, "stop")
    
    done := false
    for !done {
        selector := workflow.NewSelector(ctx)
        
        selector.AddFuture(timer.Future, func(f workflow.Future) {
            workflow.GetLogger(ctx).Info("User inactive - logging out")
            done = true
        })
        
        selector.AddReceive(userActivityChan, func(c workflow.Channel, more bool) {
            var activity struct {
                Type    string
                Timeout time.Duration
            }
            c.Receive(ctx, &activity)
            
            // Reset with possibly different duration
            if activity.Timeout > 0 {
                timeout = activity.Timeout
            }
            timer.Reset(timeout)
            
            workflow.GetLogger(ctx).Info("Activity detected", 
                "type", activity.Type, 
                "new_timeout", timeout)
        })
        
        selector.AddReceive(stopChan, func(c workflow.Channel, more bool) {
            var stop bool
            c.Receive(ctx, &stop)
            done = true
        })
        
        selector.Select(ctx)
    }
    
    return nil
}
```

#### API Reference

##### Types

```go
// Timer is the concrete implementation of a resettable timer
type Timer struct {
    Future workflow.Future  // Use this with workflow.Selector
    // contains unexported fields
}

// ResettableTimer is the interface that Timer implements
type ResettableTimer interface {
    workflow.Future
    
    // Reset cancels the current timer and starts a new one with the given duration.
    // If the timer has already fired, Reset has no effect.
    Reset(d time.Duration)
}
```

##### Functions

```go
// New creates a new resettable timer that fires after duration d.
func New(ctx workflow.Context, d time.Duration) *Timer
```

##### Usage

```go
// Reset cancels the current timer and starts a new one with the given duration
timer.Reset(30 * time.Second)

// Access the underlying Future for use with workflow.Selector
selector.AddFuture(timer.Future, func(f workflow.Future) {
    // timer fired
})

// Get blocks until the timer fires
err := timer.Future.Get(ctx, nil)

// IsReady returns true if the timer has fired
if timer.Future.IsReady() {
    // timer has fired
}
```

#### Important Notes

1. **Use with Selector**: When using the timer with `workflow.Selector`, you access the Future field directly:
   ```go
   selector.AddFuture(timer.Future, func(f workflow.Future) {
       // timer fired
   })
   ```

2. **Reset After Fire**: Once a timer has fired, calling `Reset()` has no effect. The timer is considered "done" after it fires.

3. **Determinism**: Like all workflow code, timer operations are deterministic and will replay correctly during workflow replay.

4. **Resolution**: Timer resolution is in seconds using `math.Ceil(d.Seconds())`, consistent with standard Cadence timers.

#### Testing

The resettable timer works seamlessly with Cadence's workflow test suite:

```go
func TestMyWorkflow(t *testing.T) {
    testSuite := &testsuite.WorkflowTestSuite{}
    env := testSuite.NewTestWorkflowEnvironment()
    
    // Simulate a signal being sent after 10 seconds (e.g., user interaction)
    // This would reset the timer in the workflow, preventing timeout
    env.RegisterDelayedCallback(func() {
        env.SignalWorkflow("user_activity", "click")
    }, 10*time.Second)
    
    env.ExecuteWorkflow(MyWorkflow)
    
    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())
}
```

#### Comparison with Standard Timers

**Standard Timer Pattern:**
```go
// Must manage timer cancellation and recreation manually
timerCtx, timerCancel := workflow.WithCancel(ctx)
timer := workflow.NewTimer(timerCtx, 30*time.Second)

// On activity - must cancel and recreate
timerCancel()
timerCtx, timerCancel = workflow.WithCancel(ctx)
timer = workflow.NewTimer(timerCtx, 30*time.Second)
```

**Resettable Timer Pattern:**
```go
// Simple creation and reset
timer := resettabletimer.New(ctx, 30*time.Second)

// On activity - just reset
timer.Reset(30 * time.Second)
```

The resettable timer encapsulates the cancellation and recreation logic, making timeout patterns much cleaner and easier to reason about.

