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

package util

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_MergeDicts(t *testing.T) {
	cases := []struct {
		dic1     map[string]string
		dic2     map[string]string
		expected map[string]string
	}{
		{map[string]string{"a": "1", "b": "2"}, map[string]string{"b": "3", "c": "4"}, map[string]string{"a": "1", "b": "3", "c": "4"}},
		{map[string]string{"a": "1"}, map[string]string{"a": "2"}, map[string]string{"a": "2"}},
		{map[string]string{}, map[string]string{"a": "1"}, map[string]string{"a": "1"}},
		{map[string]string{"a": "1"}, map[string]string{}, map[string]string{"a": "1"}},
		{map[string]string{}, map[string]string{}, map[string]string{}},
	}

	for _, c := range cases {
		result := MergeDicts(c.dic1, c.dic2)
		assert.Equal(t, c.expected, result)
	}
}

func Test_AwaitWaitGroup(t *testing.T) {
	cases := []struct {
		timeout  time.Duration
		expected bool
	}{
		{10 * time.Millisecond, false},
		{500 * time.Millisecond, true},
	}

	for _, c := range cases {
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			time.Sleep(100 * time.Millisecond)
			wg.Done()
		}()

		assert.Equal(t, c.expected, AwaitWaitGroup(&wg, c.timeout))
	}
}

func Test_IsTypeByteSlice(t *testing.T) {
	assert.True(t, IsTypeByteSlice(reflect.TypeOf([]byte(nil))))
	assert.True(t, IsTypeByteSlice(reflect.PtrTo(reflect.TypeOf([]byte(nil)))))
	assert.False(t, IsTypeByteSlice(reflect.TypeOf([]int(nil))))
}
