package threadsafe

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mapTestSuite is a generic test suite for the Map interface.
// It can be instantiated with different key and value types.
type mapTestSuite[K comparable, V any] struct {
	newMap func() Map[K, V]
	key1   K
	key2   K
	key3   K
	val1   V
	val2   V
	val3   V
}

func TestMutexMapImplementsMap(_ *testing.T) {
	var _ Map[string, int] = &MutexMap[string, int]{}
}

func TestSyncMapImplementsMap(_ *testing.T) {
	var _ Map[string, int] = &SyncMap[string, int]{}
}

func (s *mapTestSuite[K, V]) TestBasicOperations(t *testing.T) {
	store := s.newMap()
	assert.Equal(t, 0, store.Len())

	// Test Set and Get
	store.Set(s.key1, s.val1)
	store.Set(s.key2, s.val2)
	assert.Equal(t, 2, store.Len())

	// Test Get
	val, exists := store.Get(s.key1)
	assert.True(t, exists)
	assert.Equal(t, s.val1, val)

	// Test Get non-existent
	_, exists = store.Get(s.key3)
	assert.False(t, exists)

	// Test Delete
	store.Delete(s.key1)
	assert.Equal(t, 1, store.Len())
	_, exists = store.Get(s.key1)
	assert.False(t, exists)

	// Test Clear
	store.Clear()
	assert.Equal(t, 0, store.Len())
}

func (s *mapTestSuite[K, V]) TestCompareAndSwap(t *testing.T) {
	store := s.newMap()
	store.Set(s.key1, s.val1)

	// Successful swap
	swapped := store.CompareAndSwap(s.key1, s.val1, s.val2)
	assert.True(t, swapped)
	val, _ := store.Get(s.key1)
	assert.Equal(t, s.val2, val)

	// Failed swap (old value doesn't match)
	swapped = store.CompareAndSwap(s.key1, s.val1, s.val3)
	assert.False(t, swapped)
	val, _ = store.Get(s.key1)
	assert.Equal(t, s.val2, val) // Value should remain unchanged
}

func (s *mapTestSuite[K, V]) TestSwap(t *testing.T) {
	store := s.newMap()

	// Test swap on new key
	prev, loaded := store.Swap(s.key1, s.val1)
	assert.False(t, loaded)
	var zeroV V
	assert.Equal(t, zeroV, prev)

	// Test swap on existing key
	prev, loaded = store.Swap(s.key1, s.val2)
	assert.True(t, loaded)
	assert.Equal(t, s.val1, prev)
	val, _ := store.Get(s.key1)
	assert.Equal(t, s.val2, val)
}

func (s *mapTestSuite[K, V]) TestGetAll(t *testing.T) {
	store := s.newMap()
	store.Set(s.key1, s.val1)
	store.Set(s.key2, s.val2)

	result := store.GetAll()
	assert.Equal(t, 2, len(result))
	assert.Equal(t, s.val1, result[s.key1])
	assert.Equal(t, s.val2, result[s.key2])
}

func (s *mapTestSuite[K, V]) TestGetMany(t *testing.T) {
	store := s.newMap()
	store.Set(s.key1, s.val1)
	store.Set(s.key2, s.val2)

	// Test getting multiple keys
	result := store.GetMany([]K{s.key1, s.key3})
	assert.Equal(t, 1, len(result))
	assert.Equal(t, s.val1, result[s.key1])
	_, exists := result[s.key3]
	assert.False(t, exists)
}

func (s *mapTestSuite[K, V]) TestSetMany(t *testing.T) {
	store := s.newMap()

	// Set multiple entries at once
	entries := map[K]V{
		s.key1: s.val1,
		s.key2: s.val2,
	}
	store.SetMany(entries)

	// Verify all entries were set
	assert.Equal(t, 2, store.Len())
	val, _ := store.Get(s.key2)
	assert.Equal(t, s.val2, val)
}

func (s *mapTestSuite[K, V]) TestRange(t *testing.T) {
	store := s.newMap()
	store.Set(s.key1, s.val1)
	store.Set(s.key2, s.val2)

	// Test Range
	var count int
	store.Range(func(_ K, _ V) bool {
		count++
		return true // continue iteration
	})
	assert.Equal(t, 2, count)

	// Test early termination
	count = 0
	store.Range(func(_ K, _ V) bool {
		count++
		return false // stop after first iteration
	})
	assert.Equal(t, 1, count)
}

// runMapTestSuite runs all tests in the suite.
func runMapTestSuite[K comparable, V any](t *testing.T, s *mapTestSuite[K, V]) {
	t.Run("BasicOperations", s.TestBasicOperations)
	t.Run("CompareAndSwap", s.TestCompareAndSwap)
	t.Run("Swap", s.TestSwap)
	t.Run("GetAll", s.TestGetAll)
	t.Run("GetMany", s.TestGetMany)
	t.Run("SetMany", s.TestSetMany)
	t.Run("Range", s.TestRange)
}

// TestMapImplementations is the main test function that sets up and runs the test suites.
func TestMapImplementations(t *testing.T) {
	t.Run("string-int", func(t *testing.T) {
		t.Run("MutexMap", func(t *testing.T) {
			suite := &mapTestSuite[string, int]{
				newMap: func() Map[string, int] {
					return NewMutexMap[string](func(a, b int) bool { return a == b })
				},
				key1: "one", key2: "two", key3: "three",
				val1: 1, val2: 2, val3: 3,
			}
			runMapTestSuite(t, suite)
		})

		t.Run("RWMutexMap", func(t *testing.T) {
			suite := &mapTestSuite[string, int]{
				newMap: func() Map[string, int] {
					return NewRWMutexMap[string](func(a, b int) bool { return a == b })
				},
				key1: "one", key2: "two", key3: "three",
				val1: 1, val2: 2, val3: 3,
			}
			runMapTestSuite(t, suite)
		})

		t.Run("SyncMap", func(t *testing.T) {
			suite := &mapTestSuite[string, int]{
				newMap: func() Map[string, int] {
					return NewSyncMap[string, int]()
				},
				key1: "one", key2: "two", key3: "three",
				val1: 1, val2: 2, val3: 3,
			}
			runMapTestSuite(t, suite)
		})
	})

	type testStruct struct {
		ID   int
		Name string
	}
	t.Run("int-struct", func(t *testing.T) {
		equalFunc := func(a, b testStruct) bool { return a.ID == b.ID && a.Name == b.Name }

		t.Run("MutexMap", func(t *testing.T) {
			suite := &mapTestSuite[int, testStruct]{
				newMap: func() Map[int, testStruct] {
					return NewMutexMap[int](equalFunc)
				},
				key1: 1, key2: 2, key3: 3,
				val1: testStruct{1, "A"}, val2: testStruct{2, "B"}, val3: testStruct{3, "C"},
			}
			runMapTestSuite(t, suite)
		})

		t.Run("RWMutexMap", func(t *testing.T) {
			suite := &mapTestSuite[int, testStruct]{
				newMap: func() Map[int, testStruct] {
					return NewRWMutexMap[int](equalFunc)
				},
				key1: 1, key2: 2, key3: 3,
				val1: testStruct{1, "A"}, val2: testStruct{2, "B"}, val3: testStruct{3, "C"},
			}
			runMapTestSuite(t, suite)
		})

		// Note: SyncMap cannot be tested with non-comparable types like testStruct
		// because its CompareAndSwap relies on the `==` operator internally.
	})
}

func TestCalculateMapDiff(t *testing.T) {
	// Test empty maps
	diff := CalculateMapDiff(
		map[string]int{},
		map[string]int{},
		func(a, b int) bool { return a == b },
	)
	assert.Equal(t, 0, len(diff.AddedOrModified))
	assert.Equal(t, 0, len(diff.Removed))

	// Test map addition
	diff = CalculateMapDiff(
		map[string]int{"a": 1},
		map[string]int{},
		func(a, b int) bool { return a == b },
	)
	assert.Equal(t, 1, len(diff.AddedOrModified))
	assert.Equal(t, 0, len(diff.Removed))

	// Test map removal
	diff = CalculateMapDiff(
		map[string]int{},
		map[string]int{"a": 1},
		func(a, b int) bool { return a == b },
	)
	assert.Equal(t, 0, len(diff.AddedOrModified))
	assert.Equal(t, 1, len(diff.Removed))

	// Test map difference
	diff = CalculateMapDiff(
		map[string]int{"a": 1, "b": 2},
		map[string]int{"a": 1, "c": 3},
		func(a, b int) bool { return a == b },
	)
	assert.Equal(t, 1, len(diff.AddedOrModified))
	assert.Equal(t, 1, len(diff.Removed))
}

// A separate concurrent test is useful to control the number of goroutines precisely.
func TestConcurrentAccess(t *testing.T) {
	implementations := []struct {
		name   string
		newMap func() Map[string, int]
	}{
		{
			name: "MutexMap",
			newMap: func() Map[string, int] {
				return NewMutexMap[string](func(a, b int) bool { return a == b })
			},
		},
		{
			name: "RWMutexMap",
			newMap: func() Map[string, int] {
				return NewRWMutexMap[string](func(a, b int) bool { return a == b })
			},
		},

		{
			name: "SyncMap",
			newMap: func() Map[string, int] {
				return NewSyncMap[string, int]()
			},
		},
	}

	for _, tt := range implementations {
		t.Run(tt.name, func(t *testing.T) {
			store := tt.newMap()
			const numGoroutines = 10
			const perGoroutine = 100

			var wg sync.WaitGroup
			wg.Add(numGoroutines)

			// Concurrent writes
			for i := range numGoroutines {
				go func(goroutineID int) {
					defer wg.Done()
					for j := range perGoroutine {
						key := strconv.Itoa(goroutineID*perGoroutine + j)
						store.Set(key, goroutineID)
					}
				}(i)
			}

			// Concurrent reads
			for range numGoroutines {
				go func() {
					for j := range perGoroutine {
						store.Get(strconv.Itoa(j))
					}
				}()
			}

			wg.Wait()

			// Verify all entries were written
			assert.Equal(t, numGoroutines*perGoroutine, store.Len())

			// Verify no data races by checking all values are within expected range
			store.Range(func(_ string, value int) bool {
				assert.True(t, value >= 0 && value < numGoroutines)
				return true
			})
		})
	}
}

//
// BENCHMARKS
//

func benchmarkMap(b *testing.B, newMap func() Map[string, int]) {
	// Simple write benchmark
	b.Run("Set", func(b *testing.B) {
		store := newMap()
		b.ResetTimer()
		for b.Loop() {
			store.Set("key", 1)
		}
	})

	// Simple read benchmark
	b.Run("Get", func(b *testing.B) {
		store := newMap()
		store.Set("key", 1)
		b.ResetTimer()
		for b.Loop() {
			store.Get("key")
		}
	})

	// Concurrent workload (90% reads, 10% writes)
	b.Run("ConcurrentReadWrite", func(b *testing.B) {
		store := newMap()
		// Pre-fill the map with some data
		for i := range 1000 {
			store.Set(strconv.Itoa(i), i)
		}
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				// Generate a key to operate on
				key := strconv.Itoa(i % 1000)
				// 90% read, 10% write
				if i%10 == 0 {
					store.Set(key, i)
				} else {
					store.Get(key)
				}
				i++
			}
		})
	})
}

func BenchmarkMapImplementations(b *testing.B) {
	b.Run("MutexMap", func(b *testing.B) {
		benchmarkMap(b, func() Map[string, int] {
			return NewMutexMap[string](func(a, b int) bool { return a == b })
		})
	})

	b.Run("SyncMap", func(b *testing.B) {
		benchmarkMap(b, func() Map[string, int] {
			return NewSyncMap[string, int]()
		})
	})
}
