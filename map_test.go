package threadsafe

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncMapImplementsMap(_ *testing.T) {
	var _ Map[string, int] = NewSyncMap[string, int]()
}

func TestSyncMap_Basic(t *testing.T) {
	store := NewSyncMap[string, int]()
	assert.Equal(t, 0, store.Len())

	// Test Set and Get
	store.Set("one", 1)
	store.Set("two", 2)
	assert.Equal(t, 2, store.Len())

	// Test Get
	val, exists := store.Get("one")
	assert.True(t, exists)
	assert.Equal(t, 1, val)

	// Test Get non-existent
	_, exists = store.Get("three")
	assert.False(t, exists)

	// Test Delete
	store.Delete("one")
	assert.Equal(t, 1, store.Len())
	_, exists = store.Get("one")
	assert.False(t, exists)

	// Test Clear
	store.Clear()
	assert.Equal(t, 0, store.Len())
}

func TestSyncMap_CompareAndSwap(t *testing.T) {
	store := NewSyncMap[string, int]()
	store.Set("key", 1)

	// Successful swap
	swapped := store.CompareAndSwap("key", 1, 2)
	assert.True(t, swapped)
	val, _ := store.Get("key")
	assert.Equal(t, 2, val)

	// Failed swap (old value doesn't match)
	swapped = store.CompareAndSwap("key", 1, 3)
	assert.False(t, swapped)
	val, _ = store.Get("key")
	assert.Equal(t, 2, val) // Value should remain unchanged
}

func TestSyncMap_Swap(t *testing.T) {
	store := NewSyncMap[string, int]()

	// Test swap on new key
	prev, loaded := store.Swap("new", 1)
	assert.False(t, loaded)
	assert.Equal(t, 0, prev) // zero value for int

	// Test swap on existing key
	prev, loaded = store.Swap("new", 2)
	assert.True(t, loaded)
	assert.Equal(t, 1, prev)
	val, _ := store.Get("new")
	assert.Equal(t, 2, val)
}

func TestSyncMap_GetAll(t *testing.T) {
	store := NewSyncMap[string, int]()
	store.Set("one", 1)
	store.Set("two", 2)
	store.Set("three", 3)

	result := store.GetAll()
	assert.Equal(t, 3, len(result))
	assert.Equal(t, 1, result["one"])
	assert.Equal(t, 2, result["two"])
	assert.Equal(t, 3, result["three"])
}

func TestSyncMap_GetMany(t *testing.T) {
	store := NewSyncMap[string, int]()
	store.Set("one", 1)
	store.Set("two", 2)
	store.Set("three", 3)

	// Test getting multiple keys
	result := store.GetMany([]string{"one", "three", "missing"})
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 1, result["one"])
	assert.Equal(t, 3, result["three"])
	_, exists := result["missing"]
	assert.False(t, exists)
}

func TestSyncMap_SetMany(t *testing.T) {
	store := NewSyncMap[string, int]()

	// Set multiple entries at once
	entries := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}
	store.SetMany(entries)

	// Verify all entries were set
	assert.Equal(t, 3, store.Len())
	val, _ := store.Get("two")
	assert.Equal(t, 2, val)
}

func TestSyncMap_Range(t *testing.T) {
	store := NewSyncMap[string, int]()
	store.Set("one", 1)
	store.Set("two", 2)
	store.Set("three", 3)

	// Test Range
	var count int
	store.Range(func(_ string, _ int) bool {
		count++
		return true // continue iteration
	})
	assert.Equal(t, 3, count)

	// Test early termination
	count = 0
	store.Range(func(_ string, _ int) bool {
		count++
		return count < 2 // stop after second iteration
	})
	assert.Equal(t, 2, count)
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

func TestSyncMap_ConcurrentAccess(t *testing.T) {
	store := NewSyncMap[string, int]()
	const numGoroutines = 10
	const perGoroutine = 1000

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
