// Package threadsafe implements thread-safe operations.
package threadsafe

import (
	"iter"
	"sync"
)

// SyncMap is a thread-safe implementation of Map using sync.Map.
// Note: the internal implementation of sync.Map requires a comparable type to run the
// CompareAndSwap operation. To circumvent this, attach an equal function to the map
// upon creation.
type SyncMap[K comparable, V any] struct {
	values sync.Map
	equal  func(V, V) bool
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
	s.values.Clear()
}

// CompareAndSwap executes the compare-and-swap operation for a key.
func (s *SyncMap[K, V]) CompareAndSwap(key K, oldValue, newValue V) bool {
	current, exists := s.Get(key)
	if !exists {
		// Handle case where key doesn't exist
		return false
	}

	if s.equal != nil {
		if s.equal(current, oldValue) {
			s.values.Store(key, newValue)
			return true
		}
		return false
	}

	// Fall back on sync.Map.CompareAndSwap, which will panic if V is not comparable
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

// LoadOrStore returns the existing value for the key if present. Otherwise, it stores and returns
// the given value. The loaded result is true if the value was loaded, false if stored.
func (s *SyncMap[K, V]) LoadOrStore(key K, value V) (V, bool) {
	v, loaded := s.values.LoadOrStore(key, value)
	if !loaded {
		return value, false
	}
	return v.(V), true //nolint:revive
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
func (s *SyncMap[K, V]) LoadAndDelete(key K) (V, bool) {
	v, loaded := s.values.LoadAndDelete(key)
	if !loaded {
		var zero V
		return zero, false
	}
	return v.(V), true //nolint:revive
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

// Equals reports whether the logical content of this map and the other map is the same. Requires
// equalFn to be provided to decide how two values of type V are compared.
func (s *SyncMap[K, V]) Equals(other Map[K, V], equalFn func(a, b V) bool) bool {
	return equals(s, other, equalFn)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (s *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	s.values.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}

// All returns an iterator over key-value pairs in the map.
// The iteration order is not guaranteed to be consistent.
func (s *SyncMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		s.values.Range(func(k, v any) bool {
			return yield(k.(K), v.(V)) //nolint:revive
		})
	}
}

// Keys returns an iterator over keys in the map.
// The iteration order is not guaranteed to be consistent.
func (s *SyncMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		s.values.Range(func(k, _ any) bool {
			return yield(k.(K)) //nolint:revive
		})
	}
}

// Values returns an iterator over values in the map.
// The iteration order is not guaranteed to be consistent.
func (s *SyncMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		s.values.Range(func(_, v any) bool {
			return yield(v.(V)) //nolint:revive
		})
	}
}

// NewSyncMap creates a new instance of SyncMap. The equalFn parameter is required to
// decide how two values of type V are compared, but can be nil if V is comparable.
func NewSyncMap[K comparable, V any](equalFn func(V, V) bool) *SyncMap[K, V] {
	return &SyncMap[K, V]{
		equal: equalFn,
	}
}

// SyncMapFromMap creates a new instance of SyncMap from values in the provided map.
func SyncMapFromMap[K comparable, V any](m map[K]V, equalFn func(V, V) bool) *SyncMap[K, V] {
	newMap := NewSyncMap[K, V](equalFn)
	newMap.SetMany(m)
	return newMap
}
