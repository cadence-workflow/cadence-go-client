// Copyright (c) 2017-2020 Uber Technologies Inc.
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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/apache/thrift/lib/go/thrift"

	"go.uber.org/cadence/internal/common"
)

// encoding is capable of encoding and decoding objects
type encoding interface {
	Marshal([]interface{}) ([]byte, error)
	Unmarshal([]byte, []interface{}) error
}

// jsonEncoding encapsulates json encoding and decoding
type jsonEncoding struct {
}

// Marshal encodes an array of object into bytes
func (g jsonEncoding) Marshal(objs []interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for i, obj := range objs {
		if err := enc.Encode(obj); err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("missing argument at index %d of type %T", i, obj)
			}
			return nil, fmt.Errorf(
				"unable to encode argument: %d, %v, with json error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return buf.Bytes(), nil
}

// Unmarshal decodes a byte array into the passed in objects
func (g jsonEncoding) Unmarshal(data []byte, objs []interface{}) error {
	dec := json.NewDecoder(bytes.NewBuffer(data))
	dec.UseNumber()
	for i, obj := range objs {
		if err := dec.Decode(obj); err != nil {
			return fmt.Errorf(
				"unable to decode argument: %d, %v, with json error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return nil
}

// thriftEncoding encapsulates thrift serializer/de-serializer.
type thriftEncoding struct{}

// Marshal encodes an array of thrift into bytes
func (g thriftEncoding) Marshal(objs []interface{}) ([]byte, error) {
	var tlist []thrift.TStruct
	for i := 0; i < len(objs); i++ {
		if !common.IsThriftType(objs[i]) {
			return nil, fmt.Errorf("pointer to thrift.TStruct type is required for %v argument", i+1)
		}
		t := reflect.ValueOf(objs[i]).Interface().(thrift.TStruct)
		tlist = append(tlist, t)
	}
	return common.TListSerialize(tlist)
}

// Unmarshal decodes an array of thrift into bytes
func (g thriftEncoding) Unmarshal(data []byte, objs []interface{}) error {
	var tlist []thrift.TStruct
	for i := 0; i < len(objs); i++ {
		rVal := reflect.ValueOf(objs[i])
		if rVal.Kind() != reflect.Ptr || !common.IsThriftType(reflect.Indirect(rVal).Interface()) {
			return fmt.Errorf("pointer to pointer thrift.TStruct type is required for %v argument", i+1)
		}
		t := reflect.New(rVal.Elem().Type().Elem()).Interface().(thrift.TStruct)
		tlist = append(tlist, t)
	}

	if err := common.TListDeserialize(tlist, data); err != nil {
		return err
	}

	for i := 0; i < len(tlist); i++ {
		reflect.ValueOf(objs[i]).Elem().Set(reflect.ValueOf(tlist[i]))
	}

	return nil
}
