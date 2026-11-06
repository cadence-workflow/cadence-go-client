// Copyright (c) 2026 Uber Technologies, Inc.
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
	"fmt"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/peer"
	"go.uber.org/yarpc/peer/hostport"
	"go.uber.org/yarpc/transport/grpc"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/internal/compatibility"
)

const (
	defaultConnectionHostPort = "127.0.0.1:7833"
	// defaultConnectionServiceName is the name the Cadence frontend registers with YARPC.
	// The frontend rejects requests addressed to any other service name.
	defaultConnectionServiceName = "cadence-frontend"
	connectionCallerName         = "cadence-client"
)

type (
	// ConnectionOptions are optional parameters for NewConnection.
	ConnectionOptions struct {
		// HostPort - The gRPC address of the Cadence frontend.
		// Optional: defaulted to "127.0.0.1:7833".
		HostPort string

		// ServiceName - The service name the Cadence frontend expects on every request.
		// Only needs to be set when a proxy in front of Cadence routes on a different name.
		// Optional: defaulted to "cadence-frontend".
		ServiceName string

		// DialOptions - Options for the gRPC dialer, such as TLS credentials
		// (see DialerCredentials in go.uber.org/yarpc/transport/grpc).
		// Optional: no dial options by default.
		DialOptions []grpc.DialOption
	}

	// Connection is a gRPC connection to the Cadence frontend. It can be used anywhere a
	// workflowserviceclient.Interface is accepted, such as when creating a worker or client.
	Connection interface {
		workflowserviceclient.Interface

		// Close stops the connection. Stop any workers and clients using it first.
		Close() error
	}

	connection struct {
		workflowserviceclient.Interface
		dispatcher *yarpc.Dispatcher
	}
)

// NewConnection creates a gRPC connection to the Cadence frontend.
// The connection is established lazily, so an unreachable frontend is reported
// by the first request, not by NewConnection.
func NewConnection(options ConnectionOptions) (Connection, error) {
	hostPort := options.HostPort
	if hostPort == "" {
		hostPort = defaultConnectionHostPort
	}
	serviceName := options.ServiceName
	if serviceName == "" {
		serviceName = defaultConnectionServiceName
	}

	transport := grpc.NewTransport()
	peerChooser := peer.NewSingle(hostport.Identify(hostPort), transport.NewDialer(options.DialOptions...))
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: connectionCallerName,
		Outbounds: yarpc.Outbounds{
			serviceName: {Unary: transport.NewOutbound(peerChooser)},
		},
	})
	if err := dispatcher.Start(); err != nil {
		return nil, fmt.Errorf("start yarpc dispatcher: %w", err)
	}

	clientConfig := dispatcher.ClientConfig(serviceName)
	service := compatibility.NewThrift2ProtoAdapter(
		apiv1.NewDomainAPIYARPCClient(clientConfig),
		apiv1.NewWorkflowAPIYARPCClient(clientConfig),
		apiv1.NewWorkerAPIYARPCClient(clientConfig),
		apiv1.NewVisibilityAPIYARPCClient(clientConfig),
		apiv1.NewScheduleAPIYARPCClient(clientConfig),
	)
	conn := &connection{Interface: service, dispatcher: dispatcher}
	return conn, nil
}

func (c *connection) Close() error {
	return c.dispatcher.Stop()
}
