package internal

import (
	"errors"
	"testing"

	// "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"go.uber.org/cadence/internal/common/debug"
)

func TestBaseWorker_pollTask_no_warnLogOnShutdown(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core, zap.Development())
	worker := newBaseWorker(baseWorkerOptions{
		maxConcurrentTask:             1,
		pollerCountWithoutAutoScaling: 1,
		identity:                      "test-identity",
		pollerTracker:                 debug.NewNoopPollerTracker(),
		taskWorker:                    &testTaskWorker{},
	}, logger, tally.NoopScope, nil)

	// mock the worker started
	worker.Start()
	worker.Stop()
	worker.pollTask()

	assert.Equal(t, 0, observed.FilterMessage("poller permit acquire error").Len())
}

func TestBaseWorker_processTask_warnLogOnOtherError(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core, zap.Development())
	worker := newBaseWorker(baseWorkerOptions{
		maxConcurrentTask:             1,
		pollerCountWithoutAutoScaling: 1,
		identity:                      "test-identity",
		pollerTracker:                 debug.NewNoopPollerTracker(),
		taskWorker:                    &testTaskWorker{},
	}, logger, tally.NoopScope, nil)

	// mock the worker started
	worker.Start()
	worker.limiterContextCancel(errors.New("test error"))
	worker.pollTask()
	worker.Stop()

	assert.LessOrEqual(t, 1, observed.FilterMessage("poller permit acquire error").Len())
}

type testTaskWorker struct{}

func (t *testTaskWorker) PollTask() (interface{}, error) {
	return nil, errors.New("poll in test will fail")
}

func (t *testTaskWorker) ProcessTask(task interface{}) error {
	return nil
}
