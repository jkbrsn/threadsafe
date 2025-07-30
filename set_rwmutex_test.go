package threadsafe

import (
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRWMutexSetImplementsSet(_ *testing.T) {
	var _ Set[string] = &RWMutexSet[string]{}
}

func TestRWMutexSetBasicOperations(t *testing.T) {
	set := NewRWMutexSet[string]()

	// Test initial state
	assert.Equal(t, 0, set.Len())
	assert.False(t, set.Has("item1"))

	// Test Add
	set.Add("item1")
	assert.Equal(t, 1, set.Len())
	assert.True(t, set.Has("item1"))

	// Test Add duplicate (should not increase length)
	set.Add("item1")
	assert.Equal(t, 1, set.Len())
	assert.True(t, set.Has("item1"))

	// Test Add multiple items
	set.Add("item2")
	set.Add("item3")
	assert.Equal(t, 3, set.Len())
	assert.True(t, set.Has("item2"))
	assert.True(t, set.Has("item3"))

	// Test Remove
	set.Remove("item2")
	assert.Equal(t, 2, set.Len())
	assert.False(t, set.Has("item2"))
	assert.True(t, set.Has("item1"))
	assert.True(t, set.Has("item3"))

	// Test Remove non-existent item
	set.Remove("nonexistent")
	assert.Equal(t, 2, set.Len())

	// Test Clear
	set.Clear()
	assert.Equal(t, 0, set.Len())
	assert.False(t, set.Has("item1"))
	assert.False(t, set.Has("item3"))
}

func TestRWMutexSetSlice(t *testing.T) {
	set := NewRWMutexSet[int]()

	// Test empty set
	slice := set.Slice()
	assert.Empty(t, slice)

	// Add items
	items := []int{3, 1, 4, 1, 5, 9} // Note: duplicates should be ignored
	for _, item := range items {
		set.Add(item)
	}

	// Get slice and sort it for consistent testing
	slice = set.Slice()
	sort.Ints(slice)

	expected := []int{1, 3, 4, 5, 9}
	assert.Equal(t, expected, slice)
	assert.Equal(t, len(expected), set.Len())
}

func TestRWMutexSetRange(t *testing.T) {
	set := NewRWMutexSet[string]()

	// Test empty set
	visited := []string{}
	set.Range(func(item string) bool {
		visited = append(visited, item)
		return true
	})
	assert.Empty(t, visited)

	// Add items
	items := []string{"apple", "banana", "cherry"}
	for _, item := range items {
		set.Add(item)
	}

	// Test full iteration
	visited = []string{}
	set.Range(func(item string) bool {
		visited = append(visited, item)
		return true
	})
	sort.Strings(visited)
	sort.Strings(items)
	assert.Equal(t, items, visited)

	// Test early termination
	visited = []string{}
	count := 0
	set.Range(func(item string) bool {
		visited = append(visited, item)
		count++
		return count < 2 // Stop after 2 items
	})
	assert.Equal(t, 2, len(visited))
}

func TestRWMutexSetConcurrentAccess(t *testing.T) {
	set := NewRWMutexSet[int]()
	const numGoroutines = 100
	const numOperationsPerGoroutine = 100

	var wg sync.WaitGroup

	// Concurrent adds
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				set.Add(base*numOperationsPerGoroutine + j)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				set.Has(j)
				set.Len()
			}
		}()
	}

	wg.Wait()

	// Verify final state
	expectedLen := numGoroutines * numOperationsPerGoroutine
	assert.Equal(t, expectedLen, set.Len())

	// Verify all items are present
	for i := 0; i < expectedLen; i++ {
		assert.True(t, set.Has(i), "Item %d should be present", i)
	}
}

func TestRWMutexSetConcurrentRemoval(t *testing.T) {
	set := NewRWMutexSet[string]()
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
			set.Remove("item" + strconv.Itoa(index))
		}(i)
	}

	wg.Wait()

	// Verify set is empty
	assert.Equal(t, 0, set.Len())
	for i := 0; i < numItems; i++ {
		assert.False(t, set.Has("item"+strconv.Itoa(i)))
	}
}

func TestRWMutexSetWithDifferentTypes(t *testing.T) {
	// Test with int
	intSet := NewRWMutexSet[int]()
	intSet.Add(42)
	assert.True(t, intSet.Has(42))
	assert.Equal(t, 1, intSet.Len())

	// Test with string
	stringSet := NewRWMutexSet[string]()
	stringSet.Add("hello")
	assert.True(t, stringSet.Has("hello"))
	assert.Equal(t, 1, stringSet.Len())

	// Test with custom comparable type
	type customType struct {
		ID   int
		Name string
	}

	customSet := NewRWMutexSet[customType]()
	item := customType{ID: 1, Name: "test"}
	customSet.Add(item)
	assert.True(t, customSet.Has(item))
	assert.Equal(t, 1, customSet.Len())
}

func TestRWMutexSetSliceImmutability(t *testing.T) {
	set := NewRWMutexSet[int]()
	set.Add(1)
	set.Add(2)
	set.Add(3)

	// Get slice and modify it
	slice := set.Slice()
	originalLen := len(slice)
	slice[0] = 999 // Modify the returned slice

	// Verify original set is unchanged
	assert.Equal(t, 3, set.Len())
	assert.False(t, set.Has(999))

	// Get a new slice to verify it's not affected
	newSlice := set.Slice()
	assert.Equal(t, originalLen, len(newSlice))
	assert.NotContains(t, newSlice, 999)
}

func BenchmarkRWMutexSetAdd(b *testing.B) {
	set := NewRWMutexSet[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		set.Add(i)
	}
}

func BenchmarkRWMutexSetHas(b *testing.B) {
	set := NewRWMutexSet[int]()
	// Pre-populate
	for i := 0; i < 1000; i++ {
		set.Add(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Has(i % 1000)
	}
}

func BenchmarkRWMutexSetConcurrentReadWrite(b *testing.B) {
	set := NewRWMutexSet[int]()
	// Pre-populate
	for i := 0; i < 1000; i++ {
		set.Add(i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				set.Has(i % 1000) // Read operation
			} else {
				set.Add(i + 1000) // Write operation
			}
			i++
		}
	})
}