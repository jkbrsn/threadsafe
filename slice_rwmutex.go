// Package threadsafe implements thread-safe operations.
package threadsafe

import (
	"iter"
	"sync"
)

// RWMutexSlice is a thread-safe buffer for any type T.
type RWMutexSlice[T any] struct {
	mu   sync.RWMutex
	data []T
}

// Append appends items to the slice.
func (b *RWMutexSlice[T]) Append(item ...T) {
	b.mu.Lock()
	b.data = append(b.data, item...)
	b.mu.Unlock()
}

// Len returns the current number of items in the buffer.
func (b *RWMutexSlice[T]) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.data)
}

// Peek returns a copy of the current buffer contents without clearing.
// The returned slice is safe to read but may be stale if new items are added concurrently.
func (b *RWMutexSlice[T]) Peek() []T {
	b.mu.RLock()
	defer b.mu.RUnlock()
	copied := make([]T, len(b.data))
	copy(copied, b.data)
	return copied
}

// All returns an iterator over all items in the slice.
// The iteration order is not guaranteed to be consistent.
func (s *RWMutexSlice[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		s.mu.RLock()
		items := make([]T, 0, len(s.data))
		for _, item := range s.data {
			items = append(items, item)
		}
		s.mu.RUnlock()

		for _, item := range items {
			if !yield(item) {
				return
			}
		}
	}
}

// Flush atomically retrieves all items and clears the buffer.
// Returns a slice with the previous contents.
func (b *RWMutexSlice[T]) Flush() []T {
	b.mu.Lock()
	defer b.mu.Unlock()
	flushed := b.data
	b.data = make([]T, 0, cap(flushed))
	return flushed
}

// RWMutexSliceFromSlice creates a new RWMutexSlice from a slice.
func RWMutexSliceFromSlice[T any](slice []T) *RWMutexSlice[T] {
	newSlice := NewRWMutexSlice[T](len(slice))
	newSlice.Append(slice...)
	return newSlice
}

// NewRWMutexSlice creates a new RWMutexSlice with an optional initial capacity.
func NewRWMutexSlice[T any](initialCap int) *RWMutexSlice[T] {
	return &RWMutexSlice[T]{
		data: make([]T, 0, initialCap),
	}
}
