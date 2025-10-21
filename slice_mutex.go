// Package threadsafe implements thread-safe operations.
package threadsafe

import (
	"iter"
	"sync"
)

// MutexSlice is a thread-safe buffer for any type T.
// It allows concurrent appends and atomic flushes.
type MutexSlice[T any] struct {
	mu   sync.Mutex
	data []T
}

// Append appends items to the buffer in a thread-safe way.
func (b *MutexSlice[T]) Append(item ...T) {
	b.mu.Lock()
	b.data = append(b.data, item...)
	b.mu.Unlock()
}

// Len returns the current number of items in the buffer.
func (b *MutexSlice[T]) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.data)
}

// Peek returns a copy of the current buffer contents without clearing.
// The returned slice is safe to read but may be stale if new items are added concurrently.
func (b *MutexSlice[T]) Peek() []T {
	b.mu.Lock()
	defer b.mu.Unlock()
	copied := make([]T, len(b.data))
	copy(copied, b.data)
	return copied
}

// All returns an iterator over all items in the slice.
// The iteration order is not guaranteed to be consistent.
func (s *MutexSlice[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		s.mu.Lock()
		items := make([]T, 0, len(s.data))
		for _, item := range s.data {
			items = append(items, item)
		}
		s.mu.Unlock()

		for _, item := range items {
			if !yield(item) {
				return
			}
		}
	}
}

// Flush atomically retrieves all items and clears the buffer.
// Returns a slice with the previous contents.
func (b *MutexSlice[T]) Flush() []T {
	b.mu.Lock()
	defer b.mu.Unlock()
	flushed := b.data
	b.data = make([]T, 0, cap(flushed))
	return flushed
}

// MutexSliceFromSlice creates a new MutexSlice from a slice.
func MutexSliceFromSlice[T any](slice []T) *MutexSlice[T] {
	newSlice := NewMutexSlice[T](len(slice))
	newSlice.Append(slice...)
	return newSlice
}

// NewMutexSlice creates a new MutexSlice with an optional initial capacity.
func NewMutexSlice[T any](initialCap int) *MutexSlice[T] {
	return &MutexSlice[T]{
		data: make([]T, 0, initialCap),
	}
}
