// Package threadsafe implements thread-safe operations.
package threadsafe

import (
	"iter"
	"sync"
)

// RWMutexSlice is a thread-safe buffer for any type T, featuring concurrent appends and atomic
// flushes.
type RWMutexSlice[T any] struct {
	mu   sync.RWMutex
	data []T
}

// Append appends items to the slice.
func (s *RWMutexSlice[T]) Append(item ...T) {
	s.mu.Lock()
	s.data = append(s.data, item...)
	s.mu.Unlock()
}

// Len returns the current number of items in the slice.
func (s *RWMutexSlice[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// Peek returns a copy of the current slice contents without clearing.
// The returned slice is safe to read but may be stale if new items are added concurrently.
func (s *RWMutexSlice[T]) Peek() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copied := make([]T, len(s.data))
	copy(copied, s.data)
	return copied
}

// All returns an iterator over all items in the slice.
// The iteration order is not guaranteed to be consistent.
func (s *RWMutexSlice[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		s.mu.RLock()
		items := make([]T, 0, len(s.data))
		items = append(items, s.data...)
		s.mu.RUnlock()

		for _, item := range items {
			if !yield(item) {
				return
			}
		}
	}
}

// Flush atomically retrieves all items and clears the slice.
// Returns a slice with the previous contents.
func (s *RWMutexSlice[T]) Flush() []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	flushed := s.data
	s.data = make([]T, 0, cap(flushed))
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
