package threadsafe

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// sliceTestSuite is a generic test suite for the Slice interface.
type sliceTestSuite[T any] struct {
	newSlice func() Slice[T]
	item1    T
	item2    T
	item3    T
	items    []T
}

func TestMutexSliceImplementsSlice(_ *testing.T) {
	var _ Slice[int] = &MutexSlice[int]{}
}

func (s *sliceTestSuite[T]) TestBasicOperations(t *testing.T) {
	slice := s.newSlice()
	assert.Equal(t, 0, slice.Len())

	slice.Append(s.item1)
	slice.Append(s.item2, s.item3)
	assert.Equal(t, 3, slice.Len())

	peeked := slice.Peek()
	assert.Equal(t, 3, len(peeked))
	assert.Equal(t, s.item1, peeked[0])
	assert.Equal(t, s.item2, peeked[1])

	flushed := slice.Flush()
	assert.Equal(t, 3, len(flushed))
	assert.Equal(t, 0, slice.Len())

	// Append after flush
	slice.Append(s.item1)
	assert.Equal(t, 1, slice.Len())
}

func (s *sliceTestSuite[T]) TestPeekDoesNotMutate(t *testing.T) {
	slice := s.newSlice()
	slice.Append(s.item1, s.item2)
	peeked := slice.Peek()
	slice.Append(s.item3)
	peeked2 := slice.Peek()

	assert.Equal(t, 2, len(peeked))
	assert.Equal(t, 3, len(peeked2))
}

func (s *sliceTestSuite[T]) TestConcurrentAppend(t *testing.T) {
	slice := s.newSlice()
	const numGoroutines = 10
	const perGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(base int) {
			defer wg.Done()
			for j := range perGoroutine {
				slice.Append(s.items[(i*perGoroutine+j)%len(s.items)])
			}
		}(i)
	}
	wg.Wait()

	assert.Equal(t, numGoroutines*perGoroutine, slice.Len())

	// Ensure all values are present
	assert.Equal(t, numGoroutines*perGoroutine, len(slice.Flush()))
}

// runSliceTestSuite runs all tests in the suite.
func runSliceTestSuite[T comparable](t *testing.T, s *sliceTestSuite[T]) {
	t.Run("BasicOperations", s.TestBasicOperations)
	t.Run("PeekDoesNotMutate", s.TestPeekDoesNotMutate)
	t.Run("ConcurrentAppend", s.TestConcurrentAppend)
}

// TestSliceImplementations sets up and runs the Slice test suite.
func TestSliceImplementations(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		t.Run("MutexSlice", func(t *testing.T) {
			suite := &sliceTestSuite[string]{
				newSlice: func() Slice[string] {
					return NewMutexSlice[string](0)
				},
				item1: "apple", item2: "banana", item3: "cherry",
				items: []string{"apple", "banana", "cherry", "orange", "lime"},
			}
			runSliceTestSuite(t, suite)
		})
		t.Run("RWMutexSlice", func(t *testing.T) {
			suite := &sliceTestSuite[string]{
				newSlice: func() Slice[string] {
					return NewRWMutexSlice[string](0)
				},
				item1: "apple", item2: "banana", item3: "cherry",
				items: []string{"apple", "banana", "cherry", "orange", "lime"},
			}
			runSliceTestSuite(t, suite)
		})
	})

	t.Run("int", func(t *testing.T) {
		t.Run("MutexSlice", func(t *testing.T) {
			suite := &sliceTestSuite[int]{
				newSlice: func() Slice[int] {
					return NewMutexSlice[int](0)
				},
				item1: 1, item2: 2, item3: 3,
				items: []int{1, 2, 3, 4, 5},
			}
			runSliceTestSuite(t, suite)
		})
		t.Run("RWMutexSlice", func(t *testing.T) {
			suite := &sliceTestSuite[int]{
				newSlice: func() Slice[int] {
					return NewRWMutexSlice[int](0)
				},
				item1: 1, item2: 2, item3: 3,
				items: []int{1, 2, 3, 4, 5},
			}
			runSliceTestSuite(t, suite)
		})
	})

	type testStruct struct {
		ID   int
		Name string
	}
	t.Run("struct", func(t *testing.T) {
		t.Run("MutexSlice", func(t *testing.T) {
			suite := &sliceTestSuite[testStruct]{
				newSlice: func() Slice[testStruct] {
					return NewMutexSlice[testStruct](0)
				},
				item1: testStruct{1, "A"},
				item2: testStruct{2, "B"},
				item3: testStruct{3, "C"},
				items: []testStruct{
					{1, "A"},
					{2, "B"},
					{3, "C"},
					{4, "D"},
					{5, "E"},
				},
			}
			runSliceTestSuite(t, suite)
		})
		t.Run("RWMutexSlice", func(t *testing.T) {
			suite := &sliceTestSuite[testStruct]{
				newSlice: func() Slice[testStruct] {
					return NewRWMutexSlice[testStruct](0)
				},
				item1: testStruct{1, "A"},
				item2: testStruct{2, "B"},
				item3: testStruct{3, "C"},
				items: []testStruct{
					{1, "A"},
					{2, "B"},
					{3, "C"},
					{4, "D"},
					{5, "E"},
				},
			}
			runSliceTestSuite(t, suite)
		})
	})
}

//
// BENCHMARKS
//

func benchmarkSlice(b *testing.B, newSlice func() Slice[string]) {
	// Simple write benchmark
	b.Run("Append", func(b *testing.B) {
		slice := newSlice()
		b.ResetTimer()
		for b.Loop() {
			slice.Append("item")
		}
	})

	// Simple read benchmark
	b.Run("Peek", func(b *testing.B) {
		slice := newSlice()
		slice.Append("item")
		b.ResetTimer()
		for b.Loop() {
			slice.Peek()
		}
	})

	// Concurrent workload (90% reads, 10% writes)
	b.Run("ConcurrentReadWrite", func(b *testing.B) {
		slice := newSlice()
		// Pre-fill the set with some data
		for i := range 1000 {
			slice.Append(strconv.Itoa(i))
		}
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				// Generate an item to operate on
				item := strconv.Itoa(i % 1000)
				// 90% read, 10% write
				if i%10 == 0 {
					slice.Append(item)
				} else {
					slice.Peek()
				}
				i++
			}
		})
	})
}

func BenchmarkSliceImplementations(b *testing.B) {
	b.Run("MutexSlice", func(b *testing.B) {
		benchmarkSlice(b, func() Slice[string] {
			return NewMutexSlice[string](0)
		})
	})

	b.Run("RWMutexSlice", func(b *testing.B) {
		benchmarkSlice(b, func() Slice[string] {
			return NewRWMutexSlice[string](0)
		})
	})
}
