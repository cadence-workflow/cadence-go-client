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

package thrift

import (
	"time"

	gogo "github.com/gogo/protobuf/types"

	"go.uber.org/cadence/internal/common"
)

func boolPtr(b bool) *bool {
	return &b
}

func toDoubleValue(v *gogo.DoubleValue) *float64 {
	if v == nil {
		return nil
	}
	return &v.Value
}

func toInt64Value(v *gogo.Int64Value) *int64 {
	if v == nil {
		return nil
	}
	return common.Int64Ptr(v.Value)
}

func timeToUnixNano(t *gogo.Timestamp) *int64 {
	if t == nil {
		return nil
	}
	timestamp, err := gogo.TimestampFromProto(t)
	if err != nil {
		panic(err)
	}
	return common.Int64Ptr(timestamp.UnixNano())
}

func durationToDays(d *gogo.Duration) *int32 {
	if d == nil {
		return nil
	}
	duration, err := gogo.DurationFromProto(d)
	if err != nil {
		panic(err)
	}
	return common.Int32Ptr(int32(duration / (24 * time.Hour)))
}

func durationToSeconds(d *gogo.Duration) *int32 {
	if d == nil {
		return nil
	}
	duration, err := gogo.DurationFromProto(d)
	if err != nil {
		panic(err)
	}
	return common.Int32Ptr(int32(duration / time.Second))
}

func int32To64(v *int32) *int64 {
	if v == nil {
		return nil
	}
	return common.Int64Ptr(int64(*v))
}

type fieldSet map[string]struct{}

func newFieldSet(mask *gogo.FieldMask) fieldSet {
	if mask == nil {
		return nil
	}
	fs := map[string]struct{}{}
	for _, field := range mask.Paths {
		fs[field] = struct{}{}
	}
	return fs
}

func (fs fieldSet) isSet(field string) bool {
	if fs == nil {
		return true
	}
	_, ok := fs[field]
	return ok
}
