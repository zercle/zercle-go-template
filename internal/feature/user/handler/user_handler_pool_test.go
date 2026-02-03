// Package handler provides concurrent tests for user handler sync.Pool operations.
package handler

import (
	"sync"
	"testing"
)

// TestErrorMapPoolConcurrent tests thread safety of error map pool.
func TestErrorMapPoolConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 100
	iterations := 1000

	// Test concurrent get/put operations
	for range numGoroutines {
		wg.Go(func() {
			for range iterations {
				m := getErrorMap()
				m["field1"] = "error1"
				m["field2"] = "error2"
				m["field3"] = "error3"
				// Verify we can read back
				_ = m["field1"]
				putErrorMap(m)
			}
		})
	}

	wg.Wait()
}

// TestErrorMapPoolDataIsolation tests that pooled maps are properly cleared.
func TestErrorMapPoolDataIsolation(t *testing.T) {
	// Get a map from the pool
	m1 := getErrorMap()
	m1["field1"] = "error1"
	m1["field2"] = "error2"

	// Return it to the pool
	putErrorMap(m1)

	// Get another map - should be empty
	m2 := getErrorMap()

	// Verify the map is empty (not containing previous data)
	if len(m2) != 0 {
		t.Errorf("expected map to be empty, got %d items", len(m2))
	}

	// Clean up
	putErrorMap(m2)
}

// TestErrorMapPoolMultipleSizes tests pool with different map sizes.
func TestErrorMapPoolMultipleSizes(t *testing.T) {
	testCases := []int{0, 1, 3, 5, 10}

	for _, size := range testCases {
		m := getErrorMap()
		for i := range size {
			m[string(rune('a'+i))] = "error"
		}

		// Verify size
		if len(m) != size {
			t.Errorf("expected size %d, got %d", size, len(m))
		}

		putErrorMap(m)
	}
}

// TestErrorMapPoolParallel is a parallel version of the concurrent test.
func TestErrorMapPoolParallel(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	numGoroutines := 50
	iterations := 500

	for range numGoroutines {
		wg.Go(func() {
			for range iterations {
				m := getErrorMap()
				m["key"] = "value"
				_ = m["key"]
				putErrorMap(m)
			}
		})
	}

	wg.Wait()
}

// TestResponseBufferPoolConcurrent tests thread safety of response buffer pool.
func TestResponseBufferPoolConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 100
	iterations := 1000

	// Test concurrent get/put operations
	for range numGoroutines {
		wg.Go(func() {
			for range iterations {
				buf := getResponseBuffer()
				*buf = append(*buf, []byte(`{"test":"data"}`)...)
				// Verify we can read back
				_ = len(*buf)
				putResponseBuffer(buf)
			}
		})
	}

	wg.Wait()
}

// TestResponseBufferPoolDataIsolation tests that pooled buffers are properly reset.
func TestResponseBufferPoolDataIsolation(t *testing.T) {
	// Get a buffer from the pool
	buf1 := getResponseBuffer()
	*buf1 = append(*buf1, []byte(`{"test":"data"}`)...)

	// Return it to the pool
	putResponseBuffer(buf1)

	// Get another buffer - should be empty
	buf2 := getResponseBuffer()

	// Verify the buffer is empty
	if len(*buf2) != 0 {
		t.Errorf("expected buffer to be empty, got %d bytes", len(*buf2))
	}

	// Clean up
	putResponseBuffer(buf2)
}

// TestResponseBufferPoolParallel is a parallel version of the concurrent test.
func TestResponseBufferPoolParallel(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	numGoroutines := 50
	iterations := 500

	for range numGoroutines {
		wg.Go(func() {
			for range iterations {
				buf := getResponseBuffer()
				*buf = append(*buf, []byte(`{"test":"data"}`)...)
				putResponseBuffer(buf)
			}
		})
	}

	wg.Wait()
}

// BenchmarkErrorMapPoolWarmUp warms up the pool before benchmarks.
func BenchmarkErrorMapPoolWarmUp(b *testing.B) {
	// Warm up the pool
	for range 1000 {
		m := getErrorMap()
		m["test"] = "value"
		putErrorMap(m)
	}
}

// BenchmarkResponseBufferPoolWarmUp warms up the pool before benchmarks.
func BenchmarkResponseBufferPoolWarmUp(b *testing.B) {
	// Warm up the pool
	for range 1000 {
		buf := getResponseBuffer()
		*buf = append(*buf, []byte(`{"test":"data"}`)...)
		putResponseBuffer(buf)
	}
}
