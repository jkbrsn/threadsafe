// Package threadsafe implements thread-safe operations.
package threadsafe

import "sync"

// SyncMap is a thread-safe implementation of Map using sync.Map.
type SyncMap[K comparable, V any] struct {
	values sync.Map
}

// Get retrieves the value for the given key.
func (s *SyncMap[K, V]) Get(key K) (V, bool) {
	value, ok := s.values.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	return value.(V), true //nolint:revive
}

// Set stores a value for the given key.
func (s *SyncMap[K, V]) Set(key K, value V) {
	s.values.Store(key, value)
}

// Delete removes the key from the store.
func (s *SyncMap[K, V]) Delete(key K) {
	s.values.Delete(key)
}

// Len returns the number of items in the store.
// Note: This is an O(n) operation as sync.Map doesn't track its size.
func (s *SyncMap[K, V]) Len() int {
	count := 0
	s.values.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// Clear removes all items from the store.
func (s *SyncMap[K, V]) Clear() {
	s.values.Range(func(key, _ any) bool {
		s.values.Delete(key)
		return true
	})
}

// CompareAndSwap executes the compare-and-swap operation for a key.
func (s *SyncMap[K, V]) CompareAndSwap(key K, oldValue, newValue V) bool {
	return s.values.CompareAndSwap(key, oldValue, newValue)
}

// Swap swaps the value for a key and returns the previous value if any.
func (s *SyncMap[K, V]) Swap(key K, value V) (V, bool) {
	old, loaded := s.values.Swap(key, value)
	if !loaded {
		var zero V
		return zero, false
	}
	return old.(V), true //nolint:revive
}

// GetAll returns all key-value pairs in the store.
func (s *SyncMap[K, V]) GetAll() map[K]V {
	result := make(map[K]V)
	s.values.Range(func(key, value any) bool {
		result[key.(K)] = value.(V) //nolint:revive
		return true
	})
	return result
}

// GetMany retrieves multiple keys at once.
func (s *SyncMap[K, V]) GetMany(keys []K) map[K]V {
	result := make(map[K]V, len(keys))
	for _, key := range keys {
		if value, ok := s.Get(key); ok {
			result[key] = value
		}
	}
	return result
}

// SetMany sets multiple key-value pairs at once.
func (s *SyncMap[K, V]) SetMany(entries map[K]V) {
	for key, value := range entries {
		s.Set(key, value)
	}
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (s *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	s.values.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}

// NewSyncMap creates a new instance of SyncMap.
func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{}
}
