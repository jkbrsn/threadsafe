package threadsafe

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRWMutexSliceImplementsSlice(_ *testing.T) {
	var _ Slice[int] = &RWMutexSlice[int]{}
}

func TestRWMutexSlice_Basic(t *testing.T) {
	store := NewRWMutexSlice[int](0)
	assert.Equal(t, 0, store.Len())

	store.Append(1)
	store.Append(2, 3)
	assert.Equal(t, 3, store.Len())

	peeked := store.Peek()
	assert.Equal(t, 3, len(peeked))
	assert.Equal(t, 1, peeked[0])
	assert.Equal(t, 2, peeked[1])
	assert.Equal(t, 3, peeked[2])

	flushed := store.Flush()
	assert.Equal(t, 3, len(flushed))
	assert.Equal(t, 0, store.Len())

	// Append after flush
	store.Append(42)
	assert.Equal(t, 1, store.Len())
}

func TestRWMutexSlice_PeekDoesNotMutate(t *testing.T) {
	store := NewRWMutexSlice[string](0)
	store.Append("foo", "bar")
	peeked := store.Peek()
	store.Append("baz")
	peeked2 := store.Peek()

	assert.Equal(t, 2, len(peeked))
	assert.Equal(t, 3, len(peeked2))
}

func TestRWMutexSlice_ConcurrentAppend(t *testing.T) {
	store := NewRWMutexSlice[int](0)
	const numGoroutines = 10
	const perGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(base int) {
			defer wg.Done()
			for j := range perGoroutine {
				store.Append(base*perGoroutine + j)
			}
		}(i)
	}
	wg.Wait()

	assert.Equal(t, numGoroutines*perGoroutine, store.Len())

	// Ensure all values are present and unique
	m := make(map[int]bool)
	for _, v := range store.Flush() {
		assert.False(t, m[v])
		m[v] = true
	}
	assert.Equal(t, numGoroutines*perGoroutine, len(m))
}

func TestRWMutexSlice_FlushIsAtomic(t *testing.T) {
	store := NewRWMutexSlice[int](0)
	for i := range 10 {
		store.Append(i)
	}
	flushed := store.Flush()
	assert.Equal(t, 0, store.Len())
	assert.Equal(t, 10, len(flushed))
}
