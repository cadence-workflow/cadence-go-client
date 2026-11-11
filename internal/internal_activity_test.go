// Copyright (c) 2017-2020 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

func TestValidateRetryPolicyJitterCoefficient(t *testing.T) {
	tests := []struct {
		name    string
		jitter  *float64
		wantErr bool
	}{
		{name: "unset", jitter: nil},
		{name: "zero", jitter: common.Float64Ptr(0)},
		{name: "recommended", jitter: common.Float64Ptr(0.2)},
		{name: "one", jitter: common.Float64Ptr(1)},
		{name: "negative", jitter: common.Float64Ptr(-0.1), wantErr: true},
		{name: "greater than one", jitter: common.Float64Ptr(1.1), wantErr: true},
		{name: "NaN", jitter: common.Float64Ptr(math.NaN()), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &s.RetryPolicy{
				InitialIntervalInSeconds: common.Int32Ptr(1),
				BackoffCoefficient:       common.Float64Ptr(2),
				MaximumAttempts:          common.Int32Ptr(3),
				JitterCoefficient:        tt.jitter,
			}
			err := validateRetryPolicy(policy)
			if tt.wantErr {
				assert.EqualError(t, err, "JitterCoefficient on retry policy must be between 0 and 1")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
