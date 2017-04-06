package cadence

import (
	"github.com/uber-common/bark"
	m "github.com/uber-go/cadence-client/.gen/go/cadence"
	"github.com/uber-go/tally"
)

type (
	// Lifecycle represents objects that can be started and stopped.
	// Both activity and workflow workers implement this interface.
	Lifecycle interface {
		Stop()
		Start() error
	}

	// WorkerOptions is to configure a worker instance,
	// for example (1) the logger or any specific metrics.
	// 	       (2) Whether to heart beat for activities automatically.
	WorkerOptions interface {
		// Optional: To set the maximum concurrent activity executions this host can have.
		// default: defaultMaxConcurrentActivityExecutionSize(10k)
		SetMaxConcurrentActivityExecutionSize(size int) WorkerOptions

		// Optional: Sets the rate limiting on number of activities that can be executed.
		// This can be used to protect down stream services from flooding.
		// default: defaultMaxActivityExecutionRate(100k)
		SetMaxActivityExecutionRate(requestPerSecond float32) WorkerOptions

		// Optional: if the activities need auto heart beating for those activities
		// by the framework
		// default: false not to heartbeat.
		SetAutoHeartBeat(auto bool) WorkerOptions

		// Optional: Sets an identify that can be used to track this host for debugging.
		// default: default identity that include hostname, groupName and process ID.
		SetIdentity(identity string) WorkerOptions

		// Optional: Metrics to be reported.
		// default: no metrics.
		SetMetrics(metricsScope tally.Scope) WorkerOptions

		// Optional: Logger framework can use to log.
		// default: default logger provided.
		SetLogger(logger bark.Logger) WorkerOptions

		// Optional: Disable running workflow workers.
		// default: false
		SetDisableWorkflowWorker(disable bool) WorkerOptions

		// Optional: Disable running activity workers.
		// default: false
		SetDisableActivityWorker(disable bool) WorkerOptions
	}
)

// NewWorkerOptions returns an instance of worker options to configure.
func NewWorkerOptions() WorkerOptions {
	return NewWorkerOptionsInternal(nil)
}

// RegisterActivity - register a activity function with the framework.
// A activity takes a context and input and returns a (result, error) or just error.
// Examples:
//	func sampleActivity(ctx context.Context, input []byte) (result []byte, err error)
//	func sampleActivity(ctx context.Context, arg1 int, arg2 string) (result *customerStruct, err error)
//	func sampleActivity(ctx context.Context) (err error)
//	func sampleActivity() (result string, err error)
//	func sampleActivity(arg1 bool) (result int, err error)
//	func sampleActivity(arg1 bool) (err error)
// Serialization of all primitive types, structures is supported ... except channels, functions, variadic, unsafe pointer.
func RegisterActivity(
	activityFunc interface{},
) error {
	thImpl := getHostEnvironment()
	return thImpl.RegisterActivity(activityFunc)
}

// NewWorker creates an instance of worker for managing workflow and activity executions.
// service 	- thrift connection to the cadence server.
// groupName 	- is the name you use to identify your client worker, also
// 		  identifies group of workflow and activity implementations that are hosted by a single worker process.
// options 	-  configure any worker specific options like logger, metrics, identity.
func NewWorker(
	service m.TChanWorkflowService,
	groupName string,
	options WorkerOptions,
) Lifecycle {
	return newAggregatedWorker(service, groupName, options)
}
