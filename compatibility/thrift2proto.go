// Copyright (c) 2021 Uber Technologies, Inc.
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

package compatibility

import (
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	internal "go.uber.org/cadence/internal/compatibility"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

// AdapterClients holds the YARPC sub-clients required by NewThrift2ProtoAdapter.
// Adding new services in the future is a non-breaking change — just add a new field.
// Omit or set a field to nil if the corresponding RPCs are not needed.
type AdapterClients struct {
	Domain     apiv1.DomainAPIYARPCClient
	Workflow   apiv1.WorkflowAPIYARPCClient
	Worker     apiv1.WorkerAPIYARPCClient
	Visibility apiv1.VisibilityAPIYARPCClient
	Schedule   apiv1.ScheduleAPIYARPCClient
}

// NewThrift2ProtoAdapter creates an adapter for mapping calls from Thrift to Protobuf types.
// This is intended to be used as compatibility layer for older client version to be able to
// communicate with newer cadence server using GRPC.
func NewThrift2ProtoAdapter(clients AdapterClients) workflowserviceclient.Interface {
	return internal.NewThrift2ProtoAdapter(
		clients.Domain,
		clients.Workflow,
		clients.Worker,
		clients.Visibility,
		clients.Schedule,
	)
}
