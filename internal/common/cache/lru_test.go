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

package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLRU(t *testing.T) {
	cache := NewLRU(5)

	cache.Put("A", "Foo")
	assert.Equal(t, "Foo", cache.Get("A"))
	assert.Nil(t, cache.Get("B"))
	assert.Equal(t, 1, cache.Size())

	cache.Put("B", "Bar")
	cache.Put("C", "Cid")
	cache.Put("D", "Delt")
	assert.Equal(t, 4, cache.Size())

	assert.Equal(t, "Bar", cache.Get("B"))
	assert.Equal(t, "Cid", cache.Get("C"))
	assert.Equal(t, "Delt", cache.Get("D"))

	cache.Put("A", "Foo2")
	assert.Equal(t, "Foo2", cache.Get("A"))

	cache.Put("E", "Epsi")
	assert.Equal(t, "Epsi", cache.Get("E"))
	assert.Equal(t, "Foo2", cache.Get("A"))
	assert.Nil(t, cache.Get("B")) // Oldest, should be evicted

	// Access C, D is now LRU
	cache.Get("C")
	cache.Put("F", "Felp")
	assert.Nil(t, cache.Get("D"))

	cache.Delete("A")
	assert.Nil(t, cache.Get("A"))
}

func TestExist(t *testing.T) {
	cache := NewLRUWithInitialCapacity(5, 5)

	assert.False(t, cache.Exist("A"))

	cache.Put("A", "Foo")
	assert.True(t, cache.Exist("A"))
	assert.False(t, cache.Exist("B"))

	cache.Put("B", "Bar")
	assert.True(t, cache.Exist("B"))

	cache.Delete("A")
	assert.False(t, cache.Exist("A"))
}

func TestPutIfNotExistSuccess(t *testing.T) {
	cache := New(2, nil)

	existing, err := cache.PutIfNotExist("A", "Foo")
	assert.NoError(t, err)
	assert.Equal(t, "Foo", existing)

	existing, err = cache.PutIfNotExist("A", "Bar")
	assert.NoError(t, err)
	assert.Equal(t, "Foo", existing)

	assert.Equal(t, "Foo", cache.Get("A"))
}

func TestNoPutInPin(t *testing.T) {
	cache := New(2, &Options{
		Pin: true,
	})

	assert.Panics(t, func() {
		cache.Put("A", "Foo")
	})
}

func TestPinningTTL(t *testing.T) {
	cache := New(3, &Options{
		Pin: true,
		TTL: time.Millisecond * 100,
	}).(*lru)

	currentTime := time.UnixMilli(123)
	cache.now = func() time.Time { return currentTime }

	// Add two elements so the cache is full
	_, err := cache.PutIfNotExist("A", "Foo")
	require.NoError(t, err)
	_, err = cache.PutIfNotExist("B", "Bar")
	require.NoError(t, err)

	// Release B so it can be evicted
	cache.Release("B")

	currentTime = currentTime.Add(time.Millisecond * 300)
	assert.Equal(t, "Foo", cache.Get("A"))

	// B can be evicted, so we can add another element
	_, err = cache.PutIfNotExist("C", "Baz")
	require.NoError(t, err)

	// A cannot be evicted since it's pinned, so we can't add another element
	_, err = cache.PutIfNotExist("D", "Qux")
	assert.ErrorContains(t, err, "Cache capacity is fully occupied with pinned elements")

	// B is gone
	assert.Nil(t, cache.Get("B"))
}

func TestLRUWithTTL(t *testing.T) {
	cache := New(5, &Options{
		TTL: time.Millisecond * 100,
	}).(*lru)

	// We will capture this in the caches now function, and advance time as needed
	currentTime := time.UnixMilli(0)
	cache.now = func() time.Time { return currentTime }

	cache.Put("A", "foo")
	assert.Equal(t, "foo", cache.Get("A"))

	currentTime = currentTime.Add(time.Millisecond * 300)

	assert.Nil(t, cache.Get("A"))
	assert.Equal(t, 0, cache.Size())
}

func TestLRUCacheConcurrentAccess(t *testing.T) {
	cache := NewLRU(5)
	values := map[string]string{
		"A": "foo",
		"B": "bar",
		"C": "zed",
		"D": "dank",
		"E": "ezpz",
	}

	for k, v := range values {
		cache.Put(k, v)
	}

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			<-start

			for j := 0; j < 1000; j++ {
				cache.Get("A")
			}
		}()
	}

	close(start)
	wg.Wait()
}

func TestRemoveFunc(t *testing.T) {
	ch := make(chan bool)
	cache := New(5, &Options{
		RemovedFunc: func(i interface{}) {
			_, ok := i.(*testing.T)
			assert.True(t, ok)
			ch <- true
		},
	})

	cache.Put("testing", t)
	cache.Delete("testing")
	assert.Nil(t, cache.Get("testing"))

	timeout := time.NewTimer(time.Millisecond * 300)
	select {
	case b := <-ch:
		assert.True(t, b)
	case <-timeout.C:
		t.Error("RemovedFunc did not send true on channel ch")
	}
}

func TestRemovedFuncWithTTL(t *testing.T) {
	ch := make(chan bool)
	cache := New(5, &Options{
		TTL: time.Millisecond * 50,
		RemovedFunc: func(i interface{}) {
			_, ok := i.(*testing.T)
			assert.True(t, ok)
			ch <- true
		},
	}).(*lru)

	// We will capture this in the caches now function, and advance time as needed
	currentTime := time.UnixMilli(0)
	cache.now = func() time.Time { return currentTime }

	cache.Put("A", t)
	assert.Equal(t, t, cache.Get("A"))

	currentTime = currentTime.Add(time.Millisecond * 100)

	assert.Nil(t, cache.Get("A"))

	select {
	case b := <-ch:
		assert.True(t, b)
	case <-time.After(100 * time.Millisecond):
		t.Error("RemovedFunc did not send true on channel ch")
	}
}
