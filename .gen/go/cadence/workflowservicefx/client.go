// Code generated by thriftrw-plugin-yarpc
// @generated

package workflowservicefx

import (
	fx "go.uber.org/fx"
	yarpc "go.uber.org/yarpc"
	transport "go.uber.org/yarpc/api/transport"
	restriction "go.uber.org/yarpc/api/x/restriction"
	thrift "go.uber.org/yarpc/encoding/thrift"

	workflowserviceclient "go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
)

// Params defines the dependencies for the WorkflowService client.
type Params struct {
	fx.In

	Provider    yarpc.ClientConfig
	Restriction restriction.Checker `optional:"true"`
}

// Result defines the output of the WorkflowService client module. It provides a
// WorkflowService client to an Fx application.
type Result struct {
	fx.Out

	Client workflowserviceclient.Interface

	// We are using an fx.Out struct here instead of just returning a client
	// so that we can add more values or add named versions of the client in
	// the future without breaking any existing code.
}

// Client provides a WorkflowService client to an Fx application using the given name
// for routing.
//
//	fx.Provide(
//		workflowservicefx.Client("..."),
//		newHandler,
//	)
func Client(name string, opts ...thrift.ClientOption) interface{} {
	return func(p Params) Result {
		cc := p.Provider.ClientConfig(name)
		if namer, ok := cc.GetUnaryOutbound().(transport.Namer); ok && p.Restriction != nil {
			if err := p.Restriction.Check(thrift.Encoding, namer.TransportName()); err != nil {
				panic(err.Error())
			}
		}
		client := workflowserviceclient.New(cc, opts...)
		return Result{Client: client}
	}
}
