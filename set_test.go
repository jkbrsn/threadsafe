package threadsafe

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setTestSuite is a generic test suite for the Set interface.
// It can be instantiated with different item types.
type setTestSuite[T comparable] struct {
	newSet func() Set[T]
	item1  T
	item2  T
	item3  T
}

func TestRWMutexSetImplementsSet(_ *testing.T) {
	var _ Set[string] = &RWMutexSet[string]{}
}

func TestSyncMapSetImplementsSet(_ *testing.T) {
	var _ Set[string] = &SyncMapSet[string]{}
}

func (s *setTestSuite[T]) TestBasicOperations(t *testing.T) {
	set := s.newSet()
	assert.Equal(t, 0, set.Len())

	// Test Add
	assert.True(t, set.Add(s.item1))
	assert.True(t, set.Add(s.item2))
	assert.Equal(t, 2, set.Len())
	assert.True(t, set.Has(s.item1))
	assert.True(t, set.Has(s.item2))

	// Test Add duplicate (should not increase length)
	assert.False(t, set.Add(s.item1))
	assert.Equal(t, 2, set.Len())
	assert.True(t, set.Has(s.item1))

	// Test Has non-existent
	assert.False(t, set.Has(s.item3))

	// Test Remove
	assert.True(t, set.Delete(s.item1))
	assert.Equal(t, 1, set.Len())
	assert.False(t, set.Has(s.item1))
	assert.True(t, set.Has(s.item2))

	// Test Delete non-existent item (should not panic)
	assert.False(t, set.Delete(s.item3))
	assert.Equal(t, 1, set.Len())

	// Test Clear
	set.Clear()
	assert.Equal(t, 0, set.Len())
	assert.False(t, set.Has(s.item1))
	assert.False(t, set.Has(s.item2))
}

func (s *setTestSuite[T]) TestSlice(t *testing.T) {
	set := s.newSet()

	// Test empty set
	slice := set.Slice()
	assert.Empty(t, slice)

	// Add items
	assert.True(t, set.Add(s.item1))
	assert.True(t, set.Add(s.item2))
	assert.False(t, set.Add(s.item1)) // Duplicate should be ignored

	// Get slice
	slice = set.Slice()
	assert.Equal(t, 2, len(slice))
	assert.Equal(t, 2, set.Len())

	// Verify all items are present (order may vary)
	items := []T{s.item1, s.item2}
	for _, item := range items {
		assert.Contains(t, slice, item)
	}
}

func (s *setTestSuite[T]) TestRange(t *testing.T) {
	set := s.newSet()

	// Test empty set
	visited := []T{}
	set.Range(func(item T) bool {
		visited = append(visited, item)
		return true
	})
	assert.Empty(t, visited)

	// Add items
	set.Add(s.item1)
	set.Add(s.item2)

	// Test full iteration
	visited = []T{}
	set.Range(func(item T) bool {
		visited = append(visited, item)
		return true
	})
	assert.Equal(t, 2, len(visited))

	// Test early termination
	visited = []T{}
	count := 0
	set.Range(func(item T) bool {
		visited = append(visited, item)
		count++
		return count < 1 // Stop after 1 item
	})
	assert.Equal(t, 1, len(visited))
}

func (s *setTestSuite[T]) TestSliceImmutability(t *testing.T) {
	set := s.newSet()
	set.Add(s.item1)
	set.Add(s.item2)

	// Get slice and verify it's a copy
	slice := set.Slice()
	originalLen := len(slice)

	// Modifying the returned slice should not affect the set
	if len(slice) > 0 {
		// We can't directly modify slice elements since we don't know the zero value,
		// but we can verify the slice length doesn't change the set
		_ = append(slice, slice[0]) // Add a duplicate
	}

	// Verify original set is unchanged
	assert.Equal(t, 2, set.Len())
	assert.True(t, set.Has(s.item1))
	assert.True(t, set.Has(s.item2))

	// Get a new slice to verify it's not affected
	newSlice := set.Slice()
	assert.Equal(t, originalLen, len(newSlice))
}

func (s *setTestSuite[T]) TestAllIterator(t *testing.T) {
	set := s.newSet()
	set.Add(s.item1)
	set.Add(s.item2)

	items := collectSeq(set.All())
	assert.ElementsMatch(t, []T{s.item1, s.item2}, items)

	var calls int
	set.All()(func(_ T) bool {
		calls++
		return false
	})
	assert.Equal(t, 1, calls)

	mutating := s.newSet()
	mutating.Add(s.item1)
	mutating.Add(s.item2)
	seenOriginal := make(map[T]bool)
	mutating.All()(func(item T) bool {
		if item == s.item1 || item == s.item2 {
			seenOriginal[item] = true
		}
		if len(seenOriginal) == 1 {
			mutating.Add(s.item3)
		}
		return true
	})
	assert.True(t, seenOriginal[s.item1])
	assert.True(t, seenOriginal[s.item2])
	assert.True(t, mutating.Has(s.item3))
}

// runSetTestSuite runs all tests in the suite.
func runSetTestSuite[T comparable](t *testing.T, s *setTestSuite[T]) {
	t.Run("BasicOperations", s.TestBasicOperations)
	t.Run("Slice", s.TestSlice)
	t.Run("Range", s.TestRange)
	t.Run("SliceImmutability", s.TestSliceImmutability)
	t.Run("AllIterator", s.TestAllIterator)
}

// TestSetImplementations is the main test function that sets up and runs the test suites.
func TestSetImplementations(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		t.Run("RWMutexSet", func(t *testing.T) {
			suite := &setTestSuite[string]{
				newSet: func() Set[string] {
					return NewRWMutexSet[string]()
				},
				item1: "apple", item2: "banana", item3: "cherry",
			}
			runSetTestSuite(t, suite)
		})

		t.Run("SyncMapSet", func(t *testing.T) {
			suite := &setTestSuite[string]{
				newSet: func() Set[string] {
					return NewSyncMapSet[string]()
				},
				item1: "apple", item2: "banana", item3: "cherry",
			}
			runSetTestSuite(t, suite)
		})
	})

	t.Run("int", func(t *testing.T) {
		t.Run("RWMutexSet", func(t *testing.T) {
			suite := &setTestSuite[int]{
				newSet: func() Set[int] {
					return NewRWMutexSet[int]()
				},
				item1: 1, item2: 2, item3: 3,
			}
			runSetTestSuite(t, suite)
		})

		t.Run("SyncMapSet", func(t *testing.T) {
			suite := &setTestSuite[int]{
				newSet: func() Set[int] {
					return NewSyncMapSet[int]()
				},
				item1: 1, item2: 2, item3: 3,
			}
			runSetTestSuite(t, suite)
		})
	})

	type testStruct struct {
		ID   int
		Name string
	}
	t.Run("struct", func(t *testing.T) {
		t.Run("RWMutexSet", func(t *testing.T) {
			suite := &setTestSuite[testStruct]{
				newSet: func() Set[testStruct] {
					return NewRWMutexSet[testStruct]()
				},
				item1: testStruct{1, "A"}, item2: testStruct{2, "B"}, item3: testStruct{3, "C"},
			}
			runSetTestSuite(t, suite)
		})

		t.Run("SyncMapSet", func(t *testing.T) {
			suite := &setTestSuite[testStruct]{
				newSet: func() Set[testStruct] {
					return NewSyncMapSet[testStruct]()
				},
				item1: testStruct{1, "A"}, item2: testStruct{2, "B"}, item3: testStruct{3, "C"},
			}
			runSetTestSuite(t, suite)
		})
	})
}

// testConcurrentSetAccess tests a set implementation for concurrent access safety.
func testConcurrentSetAccess(t *testing.T, set Set[string]) {
	const numGoroutines = 10
	const perGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent writes
	for i := range numGoroutines {
		go func(goroutineID int) {
			defer wg.Done()
			for j := range perGoroutine {
				item := strconv.Itoa(goroutineID*perGoroutine + j)
				set.Add(item)
			}
		}(i)
	}

	// Concurrent reads
	for range numGoroutines {
		go func() {
			for j := range perGoroutine {
				set.Has(strconv.Itoa(j))
			}
		}()
	}

	wg.Wait()

	// Verify all entries were written
	assert.Equal(t, numGoroutines*perGoroutine, set.Len())

	// Verify no data races by checking all items are present
	for i := 0; i < numGoroutines*perGoroutine; i++ {
		assert.True(t, set.Has(strconv.Itoa(i)), "Item %d should be present", i)
	}
}

// A separate concurrent test is useful to control the number of goroutines precisely.
func TestSetConcurrentAccess(t *testing.T) {
	implementations := []struct {
		name   string
		newSet func() Set[string]
	}{
		{
			name: "RWMutexSet",
			newSet: func() Set[string] {
				return NewRWMutexSet[string]()
			},
		},
		{
			name: "SyncMapSet",
			newSet: func() Set[string] {
				return NewSyncMapSet[string]()
			},
		},
	}

	for _, tt := range implementations {
		t.Run(tt.name, func(t *testing.T) {
			testConcurrentSetAccess(t, tt.newSet())
		})
	}
}

func TestSetConcurrentRemoval(t *testing.T) {
	implementations := []struct {
		name   string
		newSet func() Set[string]
	}{
		{
			name: "RWMutexSet",
			newSet: func() Set[string] {
				return NewRWMutexSet[string]()
			},
		},
		{
			name: "SyncMapSet",
			newSet: func() Set[string] {
				return NewSyncMapSet[string]()
			},
		},
	}

	for _, tt := range implementations {
		t.Run(tt.name, func(t *testing.T) {
			set := tt.newSet()
			const numItems = 1000

			// Pre-populate the set
			for i := 0; i < numItems; i++ {
				set.Add("item" + strconv.Itoa(i))
			}

			var wg sync.WaitGroup

			// Concurrent removals
			for i := 0; i < numItems; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					set.Delete("item" + strconv.Itoa(index))
				}(i)
			}

			wg.Wait()

			// Verify set is empty
			assert.Equal(t, 0, set.Len())
			for i := 0; i < numItems; i++ {
				assert.False(t, set.Has("item"+strconv.Itoa(i)))
			}
		})
	}
}

//
// BENCHMARKS
//

func benchmarkSet(b *testing.B, newSet func() Set[string]) {
	// Simple write benchmark
	b.Run("Add", func(b *testing.B) {
		set := newSet()
		b.ResetTimer()
		for b.Loop() {
			set.Add("item")
		}
	})

	// Simple read benchmark
	b.Run("Has", func(b *testing.B) {
		set := newSet()
		set.Add("item")
		b.ResetTimer()
		for b.Loop() {
			set.Has("item")
		}
	})

	// Concurrent workload (90% reads, 10% writes)
	b.Run("ConcurrentReadWrite", func(b *testing.B) {
		set := newSet()
		// Pre-fill the set with some data
		for i := range 1000 {
			set.Add(strconv.Itoa(i))
		}
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				// Generate an item to operate on
				item := strconv.Itoa(i % 1000)
				// 90% read, 10% write
				if i%10 == 0 {
					set.Add(item)
				} else {
					set.Has(item)
				}
				i++
			}
		})
	})
}

func BenchmarkSetImplementations(b *testing.B) {
	b.Run("RWMutexSet", func(b *testing.B) {
		benchmarkSet(b, func() Set[string] {
			return NewRWMutexSet[string]()
		})
	})

	b.Run("SyncMapSet", func(b *testing.B) {
		benchmarkSet(b, func() Set[string] {
			return NewSyncMapSet[string]()
		})
	})
}
