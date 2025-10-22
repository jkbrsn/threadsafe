// Package threadsafe implements thread-safe operations.
package threadsafe

import (
	"iter"
	"sync"
)

// MutexSlice is a thread-safe buffer for any type T, featuring concurrent appends and atomic
// flushes.
type MutexSlice[T any] struct {
	mu   sync.Mutex
	data []T
}

// Append appends items to the slice in a thread-safe way.
func (s *MutexSlice[T]) Append(item ...T) {
	s.mu.Lock()
	s.data = append(s.data, item...)
	s.mu.Unlock()
}

// Len returns the current number of items in the slice.
func (s *MutexSlice[T]) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.data)
}

// Peek returns a copy of the current slice contents without clearing.
// The returned slice is safe to read but may be stale if new items are added concurrently.
func (s *MutexSlice[T]) Peek() []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := make([]T, len(s.data))
	copy(copied, s.data)
	return copied
}

// All returns an iterator over all items in the slice.
// The iteration order is not guaranteed to be consistent.
func (s *MutexSlice[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		s.mu.Lock()
		items := make([]T, 0, len(s.data))
		items = append(items, s.data...)
		s.mu.Unlock()

		for _, item := range items {
			if !yield(item) {
				return
			}
		}
	}
}

// Flush atomically retrieves all items and clears the slice.
// Returns a slice with the previous contents.
func (s *MutexSlice[T]) Flush() []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	flushed := s.data
	s.data = make([]T, 0, cap(flushed))
	return flushed
}

// MutexSliceFromSlice creates a new MutexSlice from a standard slice.
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
