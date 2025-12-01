package threadsafe

import (
	"maps"
	"slices"
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
	equal  func(a, b V) bool
}

func TestMutexMapImplementsMap(_ *testing.T) {
	var _ Map[string, int] = &MutexMap[string, int]{}
}

func TestRWMutexMapImplementsMap(_ *testing.T) {
	var _ Map[string, int] = &RWMutexMap[string, int]{}
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

func (s *mapTestSuite[K, V]) TestLoadOrStore(t *testing.T) {
	store := s.newMap()

	v, loaded := store.LoadOrStore(s.key1, s.val1)
	assert.False(t, loaded)
	assert.Equal(t, s.val1, v)

	v, loaded = store.LoadOrStore(s.key1, s.val2)
	assert.True(t, loaded)
	assert.Equal(t, s.val1, v)

	assert.Equal(t, 1, store.Len())
}

func (s *mapTestSuite[K, V]) TestLoadAndDelete(t *testing.T) {
	store := s.newMap()

	// Non-existent key
	v, loaded := store.LoadAndDelete(s.key1)
	var zeroV V
	assert.False(t, loaded)
	assert.Equal(t, zeroV, v)
	assert.Equal(t, 0, store.Len())

	// Existing key
	store.Set(s.key1, s.val1)
	v, loaded = store.LoadAndDelete(s.key1)
	assert.True(t, loaded)
	assert.Equal(t, s.val1, v)
	_, ok := store.Get(s.key1)
	assert.False(t, ok)
	assert.Equal(t, 0, store.Len())
}

func (s *mapTestSuite[K, V]) TestIterators(t *testing.T) {
	require := assert.New(t)

	store := s.newMap()
	store.Set(s.key1, s.val1)
	store.Set(s.key2, s.val2)

	expected := store.GetAll()

	// All yields every key/value pair exactly once for the current snapshot.
	k, v := collectSeq2(store.All())
	require.Len(k, len(expected))
	require.Len(v, len(expected))
	seen := make(map[K]struct{}, len(expected))
	for i := range k {
		val, ok := expected[k[i]]
		require.True(ok)
		require.True(s.equal(val, v[i]))
		seen[k[i]] = struct{}{}
	}
	require.Len(seen, len(expected))

	// All respects early termination via yield returning false.
	var allCalls int
	store.All()(func(_ K, _ V) bool {
		allCalls++
		return false
	})
	require.Equal(1, allCalls)

	// Keys iterator matches expected keys and respects early termination.
	keys := collectSeq(store.Keys())
	require.Len(keys, len(expected))
	for _, key := range keys {
		_, ok := expected[key]
		require.True(ok)
	}

	var keyCalls int
	store.Keys()(func(_ K) bool {
		keyCalls++
		return false
	})
	require.Equal(1, keyCalls)

	// Values iterator matches expected values and respects early termination.
	expectedValues := make([]V, 0, len(expected))
	for _, val := range expected {
		expectedValues = append(expectedValues, val)
	}
	values := collectSeq(store.Values())
	require.Len(values, len(expectedValues))
	used := make([]bool, len(expectedValues))
	for _, got := range values {
		found := false
		for i, exp := range expectedValues {
			if used[i] {
				continue
			}
			if s.equal(exp, got) {
				used[i] = true
				found = true
				break
			}
		}
		require.True(found)
	}

	var valueCalls int
	store.Values()(func(_ V) bool {
		valueCalls++
		return false
	})
	require.Equal(1, valueCalls)

	// Iteration remains safe when mutating during traversal: original keys are still observed.
	mutating := s.newMap()
	mutating.Set(s.key1, s.val1)
	mutating.Set(s.key2, s.val2)
	before := mutating.GetAll()
	seenDuringMutation := make(map[K]bool, len(before))
	mutating.All()(func(key K, value V) bool {
		if exp, ok := before[key]; ok {
			require.True(s.equal(exp, value))
			seenDuringMutation[key] = true
		}
		if len(seenDuringMutation) == 1 {
			mutating.Set(s.key3, s.val3)
		}
		return true
	})
	for k := range before {
		require.True(seenDuringMutation[k])
	}
	require.Equal(len(before)+1, mutating.Len())
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
	t.Run("LoadOrStore", s.TestLoadOrStore)
	t.Run("LoadAndDelete", s.TestLoadAndDelete)
	if s.equal != nil {
		t.Run("Iterators", s.TestIterators)
	}
}

// testStringIntMapImplementations tests all map implementations with string-int types.
func testStringIntMapImplementations(t *testing.T) {
	t.Run("MutexMap", func(t *testing.T) {
		suite := &mapTestSuite[string, int]{
			newMap: func() Map[string, int] {
				return NewMutexMap[string](func(a, b int) bool { return a == b })
			},
			key1: "one", key2: "two", key3: "three",
			val1: 1, val2: 2, val3: 3,
			equal: func(a, b int) bool { return a == b },
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
			equal: func(a, b int) bool { return a == b },
		}
		runMapTestSuite(t, suite)
	})

	t.Run("SyncMap", func(t *testing.T) {
		suite := &mapTestSuite[string, int]{
			newMap: func() Map[string, int] {
				return NewSyncMap[string](func(a, b int) bool { return a == b })
			},
			key1: "one", key2: "two", key3: "three",
			val1: 1, val2: 2, val3: 3,
			equal: func(a, b int) bool { return a == b },
		}
		runMapTestSuite(t, suite)
	})

	t.Run("SyncMap (nil equalFn for comparable V)", func(t *testing.T) {
		suite := &mapTestSuite[string, int]{
			newMap: func() Map[string, int] {
				return NewSyncMap[string, int](nil)
			},
			key1: "one", key2: "two", key3: "three",
			val1: 1, val2: 2, val3: 3,
			equal: func(a, b int) bool { return a == b },
		}
		runMapTestSuite(t, suite)
	})
}

// testIntStructMapImplementations tests all map implementations with int-struct types.
func testIntStructMapImplementations(t *testing.T) {
	type testStruct struct {
		ID   int
		Name string
	}
	equalFunc := func(a, b testStruct) bool { return a.ID == b.ID && a.Name == b.Name }

	t.Run("MutexMap", func(t *testing.T) {
		suite := &mapTestSuite[int, testStruct]{
			newMap: func() Map[int, testStruct] {
				return NewMutexMap[int](equalFunc)
			},
			key1: 1, key2: 2, key3: 3,
			val1: testStruct{1, "A"}, val2: testStruct{2, "B"}, val3: testStruct{3, "C"},
			equal: equalFunc,
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
			equal: equalFunc,
		}
		runMapTestSuite(t, suite)
	})

	t.Run("SyncMap", func(t *testing.T) {
		suite := &mapTestSuite[int, testStruct]{
			newMap: func() Map[int, testStruct] {
				return NewSyncMap[int](equalFunc)
			},
			key1: 1, key2: 2, key3: 3,
			val1: testStruct{1, "A"}, val2: testStruct{2, "B"}, val3: testStruct{3, "C"},
			equal: equalFunc,
		}
		runMapTestSuite(t, suite)
	})
}

// TestMapImplementations is the main test function that sets up and runs the test suites.
func TestMapImplementations(t *testing.T) {
	t.Run("string-int", testStringIntMapImplementations)
	t.Run("int-struct", testIntStructMapImplementations)
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

// testConcurrentMapAccess tests a map implementation for concurrent access safety.
func testConcurrentMapAccess(t *testing.T, store Map[string, int]) {
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
				return NewSyncMap[string](func(a, b int) bool { return a == b })
			},
		},
	}

	for _, tt := range implementations {
		t.Run(tt.name, func(t *testing.T) {
			testConcurrentMapAccess(t, tt.newMap())
		})
	}
}

func TestMapZeroValue(t *testing.T) {
	t.Run("RWMutexMap", func(t *testing.T) {
		var m RWMutexMap[string, int]

		// Set on zero-value should initialize map
		m.Set("key1", 1)
		m.Set("key2", 2)
		assert.Equal(t, 2, m.Len())

		// Get should work
		val, ok := m.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, 1, val)

		// Delete should work
		m.Delete("key1")
		assert.Equal(t, 1, m.Len())

		// Read operations on zero-value
		var m2 RWMutexMap[int, string]
		_, ok = m2.Get(999)
		assert.False(t, ok)
		assert.Equal(t, 0, m2.Len())

		// Delete on zero-value with nil map
		var m3 RWMutexMap[string, int]
		m3.Delete("anything") // Should not panic
		assert.Equal(t, 0, m3.Len())
	})

	t.Run("MutexMap", func(t *testing.T) {
		var m MutexMap[string, int]

		// Set on zero-value should initialize map
		m.Set("key1", 1)
		m.Set("key2", 2)
		assert.Equal(t, 2, m.Len())

		// Get should work
		val, ok := m.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, 1, val)

		// Delete should work
		m.Delete("key1")
		assert.Equal(t, 1, m.Len())

		// Read operations on zero-value
		var m2 MutexMap[int, string]
		_, ok = m2.Get(999)
		assert.False(t, ok)
		assert.Equal(t, 0, m2.Len())

		// Delete on zero-value with nil map
		var m3 MutexMap[string, int]
		m3.Delete("anything") // Should not panic
		assert.Equal(t, 0, m3.Len())
	})

	t.Run("SyncMap", func(t *testing.T) {
		// SyncMap is already zero-value safe (sync.Map is zero-value safe)
		var m SyncMap[string, int]

		// Set on zero-value
		m.Set("key1", 1)
		m.Set("key2", 2)
		assert.Equal(t, 2, m.Len())

		// Get should work
		val, ok := m.Get("key1")
		assert.True(t, ok)
		assert.Equal(t, 1, val)

		// Read operations work on zero-value
		var m2 SyncMap[int, string]
		_, ok = m2.Get(999)
		assert.False(t, ok)
		assert.Equal(t, 0, m2.Len())
	})
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

	b.Run("RWMutexMap", func(b *testing.B) {
		benchmarkMap(b, func() Map[string, int] {
			return NewRWMutexMap[string](func(a, b int) bool { return a == b })
		})
	})

	b.Run("SyncMap", func(b *testing.B) {
		benchmarkMap(b, func() Map[string, int] {
			return NewSyncMap[string](func(a, b int) bool { return a == b })
		})
	})
}

func BenchmarkMapIterationPatterns(b *testing.B) {
	const size = 1024
	keys := make([]string, size)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}

	run := func(b *testing.B, name string, newMap func() Map[string, int]) {
		b.Run(name, func(b *testing.B) {
			store := newMap()
			for i, key := range keys {
				store.Set(key, i)
			}
			b.ReportAllocs()

			b.Run("AllForRange", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					count := 0
					for range store.All() {
						count++
					}
					if count != size {
						b.Fatalf("unexpected count: %d", count)
					}
				}
			})

			b.Run("RangeCallback", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					count := 0
					store.Range(func(_ string, _ int) bool {
						count++
						return true
					})
					if count != size {
						b.Fatalf("unexpected count: %d", count)
					}
				}
			})

			type kv struct {
				k string
				v int
			}

			b.Run("CollectManual", func(b *testing.B) {
				b.ReportAllocs()
				buf := make([]kv, 0, size)
				for b.Loop() {
					buf = buf[:0]
					for k, v := range store.All() {
						buf = append(buf, kv{k: k, v: v})
					}
					if len(buf) != size {
						b.Fatalf("unexpected len: %d", len(buf))
					}
				}
			})

			b.Run("CollectMapsCollect", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					out := maps.Collect(store.All())
					if len(out) != size {
						b.Fatalf("unexpected len: %d", len(out))
					}
				}
			})

			b.Run("CollectValuesSlice", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					out := slices.Collect(store.Values())
					if len(out) != size {
						b.Fatalf("unexpected len: %d", len(out))
					}
				}
			})
		})
	}

	run(b, "MutexMap", func() Map[string, int] {
		return NewMutexMap[string](func(a, b int) bool { return a == b })
	})
	run(b, "RWMutexMap", func() Map[string, int] {
		return NewRWMutexMap[string](func(a, b int) bool { return a == b })
	})
	run(b, "SyncMap", func() Map[string, int] {
		return NewSyncMap[string](func(a, b int) bool { return a == b })
	})
}

func BenchmarkSyncMapClear(b *testing.B) {
	const size = 2048
	keys := make([]string, size)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}

	clearWithRangeDelete := func(s *SyncMap[string, int]) {
		s.values.Range(func(k, _ any) bool {
			s.values.Delete(k)
			return true
		})
	}

	benchmark := func(b *testing.B, name string, clearFn func(*SyncMap[string, int])) {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				s := NewSyncMap[string, int](nil)
				for i, key := range keys {
					s.Set(key, i)
				}
				clearFn(s)
				if s.Len() != 0 {
					b.Fatal("map not cleared")
				}
			}
		})
	}

	benchmark(b, "NativeClear", func(s *SyncMap[string, int]) {
		s.Clear()
	})
	benchmark(b, "RangeDelete", clearWithRangeDelete)
}
