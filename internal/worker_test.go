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

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	EnableVerboseLogging(true)
	m.Run()
}

func Test_NewWorker(t *testing.T) {
	tests := []struct {
		name      string
		options   WorkerOptions
		expectErr string
	}{
		{
			name:      "happy with default value",
			options:   WorkerOptions{},
			expectErr: "",
		},
		{
			name: "happy with explicit decision task poller set to 1 if sticky task list is disabled",
			options: WorkerOptions{
				MaxConcurrentDecisionTaskPollers: 1,
				DisableStickyExecution:           true,
			},
			expectErr: "",
		},
		{
			name: "invalid worker with explicit decision task poller set to 1",
			options: WorkerOptions{
				MaxConcurrentDecisionTaskPollers: 1,
			},
			expectErr: "DecisionTaskPollers must be >= 2 or use default value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := NewWorker(nil, "test-domain", "test-tasklist", tt.options)
			if tt.expectErr != "" {
				assert.ErrorContains(t, err, tt.expectErr)
				assert.Nil(t, w)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, w)
			}
		})
	}
}
