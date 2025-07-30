// Package threadsafe implements thread-safe operations.
package threadsafe

import "sync"

// SyncMapSet is a thread-safe Set implementation backed by sync.Map.
// Internally it stores the items as keys in the sync.Map with an empty struct{} value.
// All operations are safe for concurrent use by multiple goroutines.
//
// NOTE: Operations like Len, Slice and Clear iterate over the entire map. They
// run in O(n) time and may allocate, just like their RWMutex counterpart.
// If you need high-frequency Len or Slice under heavy write load, prefer the
// RWMutex variant which can maintain a separate atomic counter.
type SyncMapSet[T comparable] struct {
	items sync.Map // key T -> struct{}
}

// NewSyncMapSet creates a new instance of SyncMapSet.
func NewSyncMapSet[T comparable]() *SyncMapSet[T] {
	return &SyncMapSet[T]{}
}

// Add stores an item in the set.
func (s *SyncMapSet[T]) Add(item T) {
	s.items.Store(item, struct{}{})
}

// Remove deletes an item from the set.
func (s *SyncMapSet[T]) Remove(item T) {
	s.items.Delete(item)
}

// Has returns true if the item is in the set, otherwise false.
func (s *SyncMapSet[T]) Has(item T) bool {
	_, exists := s.items.Load(item)
	return exists
}

// Len returns the number of items in the set.
func (s *SyncMapSet[T]) Len() int {
	count := 0
	s.items.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// Clear removes all items from the set.
func (s *SyncMapSet[T]) Clear() {
	s.items.Range(func(key, _ any) bool {
		s.items.Delete(key)
		return true
	})
}

// Slice returns a copy of the set as a slice.
func (s *SyncMapSet[T]) Slice() []T {
	result := make([]T, 0)
	s.items.Range(func(key, _ any) bool {
		result = append(result, key.(T))
		return true
	})
	return result
}

// Range calls f sequentially for each item present in the set.
// If f returns false, range stops the iteration.
func (s *SyncMapSet[T]) Range(f func(item T) bool) {
	s.items.Range(func(key, _ any) bool {
		return f(key.(T))
	})
}
