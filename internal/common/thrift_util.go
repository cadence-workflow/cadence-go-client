// Copyright (c) 2017 Uber Technologies, Inc.
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

package common

import (
	"context"
	"reflect"

	"github.com/apache/thrift/lib/go/thrift"
)

// TListSerialize is used to serialize list of thrift TStruct to []byte
func TListSerialize(ts []thrift.TStruct) ([]byte, error) {
	if ts == nil {
		return nil, nil
	}

	t := thrift.NewTSerializer()

	// NOTE: we don't write any markers as thrift by design being a streaming protocol doesn't
	// recommend writing length.

	// TODO populate context from argument
	ctx := context.Background()
	for _, v := range ts {
		if e := v.Write(ctx, t.Protocol); e != nil {
			return nil, thrift.PrependError("error writing TStruct: ", e)
		}
	}

	return t.Transport.Bytes(), t.Protocol.Flush(ctx)
}

// TListDeserialize is used to deserialize []byte to list of thrift TStruct
func TListDeserialize(ts []thrift.TStruct, b []byte) (err error) {
	t := thrift.NewTDeserializer()
	err = nil
	if _, err = t.Transport.Write(b); err != nil {
		return
	}

	// TODO populate context from argument
	ctx := context.Background()
	for i := 0; i < len(ts); i++ {
		if e := ts[i].Read(ctx, t.Protocol); e != nil {
			err = thrift.PrependError("error reading TStruct: ", e)
			return
		}
	}

	return
}

// IsUseThriftEncoding checks if the objects passed in are all encoded using thrift.
func IsUseThriftEncoding(objs []interface{}) bool {
	if len(objs) == 0 {
		return false
	}
	// NOTE: our criteria to use which encoder is simple if all the types are serializable using thrift then we use
	// thrift encoder. For everything else we default to gob.
	for _, obj := range objs {
		if !IsThriftType(obj) {
			return false
		}
	}
	return true
}

// IsUseThriftDecoding checks if the objects passed in are all de-serializable using thrift.
func IsUseThriftDecoding(objs []interface{}) bool {
	if len(objs) == 0 {
		return false
	}
	// NOTE: our criteria to use which encoder is simple if all the types are de-serializable using thrift then we use
	// thrift decoder. For everything else we default to gob.
	for _, obj := range objs {
		rVal := reflect.ValueOf(obj)
		if rVal.Kind() != reflect.Ptr || !IsThriftType(reflect.Indirect(rVal).Interface()) {
			return false
		}
	}
	return true
}

// IsThriftType checks whether the object passed in is a thrift encoded object.
func IsThriftType(v interface{}) bool {
	// NOTE: Thrift serialization works only if the values are pointers.
	// Thrift has a validation that it meets thift.TStruct which has Read/Write pointer receivers.

	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return false
	}
	return reflect.TypeOf(v).Implements(tStructType)
}

var tStructType = reflect.TypeOf((*thrift.TStruct)(nil)).Elem()
