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
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/grpc"

	"go.uber.org/cadence/.gen/go/shared"
)

type fakeDomainAPI struct {
	apiv1.DomainAPIYARPCServer
}

func (fakeDomainAPI) DescribeDomain(_ context.Context, request *apiv1.DescribeDomainRequest) (*apiv1.DescribeDomainResponse, error) {
	response := &apiv1.DescribeDomainResponse{Domain: &apiv1.Domain{Name: request.GetName()}}
	return response, nil
}

func TestNewConnection(t *testing.T) {
	tests := []struct {
		name                string
		frontendServiceName string
		serviceNameOption   string
	}{
		{
			name:                "default service name reaches cadence-frontend",
			frontendServiceName: "cadence-frontend",
		},
		{
			name:                "custom service name",
			frontendServiceName: "cadence-proxy",
			serviceNameOption:   "cadence-proxy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostPort := startFakeFrontend(t, tt.frontendServiceName)

			conn, err := NewConnection(ConnectionOptions{HostPort: hostPort, ServiceName: tt.serviceNameOption})
			require.NoError(t, err)
			t.Cleanup(func() {
				err := conn.Close()
				assert.NoError(t, err)
			})

			domainName := "test-domain"
			// YARPC rejects outbound calls without a deadline; the timeout is only a safety net.
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			response, err := conn.DescribeDomain(ctx, &shared.DescribeDomainRequest{Name: &domainName})
			require.NoError(t, err)
			assert.Equal(t, domainName, response.GetDomainInfo().GetName())
		})
	}
}

// startFakeFrontend serves the domain API under serviceName on a local port and returns its address.
func startFakeFrontend(t *testing.T, serviceName string) string {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name:     serviceName,
		Inbounds: yarpc.Inbounds{grpc.NewTransport().NewInbound(listener)},
	})
	dispatcher.Register(apiv1.BuildDomainAPIYARPCProcedures(fakeDomainAPI{}))
	err = dispatcher.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		err := dispatcher.Stop()
		assert.NoError(t, err)
	})

	return listener.Addr().String()
}
