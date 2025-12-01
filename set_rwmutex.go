// Package threadsafe implements thread-safe operations.
package threadsafe

import (
	"iter"
	"sync"
)

// RWMutexSet is a thread-safe implementation of Set using sync.RWMutex.
type RWMutexSet[T comparable] struct {
	mu    sync.RWMutex
	items map[T]struct{}
	size  int // Separate size counter for O(1) Len
}

// Add stores an item in the set.
func (s *RWMutexSet[T]) Add(item T) (added bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.items == nil {
		s.items = make(map[T]struct{})
	}

	if _, exists := s.items[item]; !exists {
		s.items[item] = struct{}{}
		s.size++
		return true
	}
	return false
}

// Delete removes an item from the set.
func (s *RWMutexSet[T]) Delete(item T) (removed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.items == nil {
		return false
	}

	if _, exists := s.items[item]; exists {
		delete(s.items, item)
		s.size--
		return true
	}
	return false
}

// Has returns true if the item is in the set, otherwise false.
func (s *RWMutexSet[T]) Has(item T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.items[item]
	return exists
}

// Len returns the number of items in the set.
func (s *RWMutexSet[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.size
}

// Clear removes all items from the set.
func (s *RWMutexSet[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(map[T]struct{})
	s.size = 0
}

// Slice returns a copy of the set as a slice.
func (s *RWMutexSet[T]) Slice() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]T, 0, len(s.items))
	for item := range s.items {
		result = append(result, item)
	}
	return result
}

// Range calls f sequentially for each item present in the set.
// If f returns false, range stops the iteration.
func (s *RWMutexSet[T]) Range(f func(item T) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for item := range s.items {
		if !f(item) {
			break
		}
	}
}

// All returns an iterator over all items in the set.
// The iteration order is not guaranteed to be consistent.
func (s *RWMutexSet[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		s.mu.RLock()
		items := make([]T, 0, len(s.items))
		for item := range s.items {
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

// NewRWMutexSet creates a new instance of RWMutexSet.
func NewRWMutexSet[T comparable]() *RWMutexSet[T] {
	return &RWMutexSet[T]{
		items: make(map[T]struct{}),
		size:  0,
	}
}
