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

package proto

import (
	"time"

	gogo "github.com/gogo/protobuf/types"

	"go.uber.org/cadence/internal/common"
)

func fromDoubleValue(v *float64) *gogo.DoubleValue {
	if v == nil {
		return nil
	}
	return &gogo.DoubleValue{Value: *v}
}

func fromInt64Value(v *int64) *gogo.Int64Value {
	if v == nil {
		return nil
	}
	return &gogo.Int64Value{Value: *v}
}

func unixNanoToTime(t *int64) *gogo.Timestamp {
	if t == nil {
		return nil
	}
	time, err := gogo.TimestampProto(time.Unix(0, *t))
	if err != nil {
		panic(err)
	}
	return time
}

func daysToDuration(d *int32) *gogo.Duration {
	if d == nil {
		return nil
	}
	return gogo.DurationProto(time.Duration(*d) * (24 * time.Hour))
}

func secondsToDuration(d *int32) *gogo.Duration {
	if d == nil {
		return nil
	}
	return gogo.DurationProto(time.Duration(*d) * time.Second)
}

func int64To32(v *int64) *int32 {
	if v == nil {
		return nil
	}
	return common.Int32Ptr(int32(*v))
}

func newFieldMask(fields []string) *gogo.FieldMask {
	return &gogo.FieldMask{Paths: fields}
}
