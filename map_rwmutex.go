// Package threadsafe implements thread-safe operations.
package threadsafe

import (
	"iter"
	"maps"
	"sync"
)

// RWMutexMap is a thread-safe implementation of Map using sync.RWMutex.
type RWMutexMap[K comparable, V any] struct {
	mu     sync.RWMutex
	values map[K]V

	equal func(V, V) bool
}

// Get retrieves the value for the given key.
func (m *RWMutexMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.values[key]
	return value, ok
}

// Set stores a value for the given key.
func (m *RWMutexMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.values[key] = value
}

// Delete removes the key from the map.
func (m *RWMutexMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.values, key)
}

// Len returns the number of items in the map.
func (m *RWMutexMap[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.values)
}

// Clear removes all items from the map.
func (m *RWMutexMap[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.values = make(map[K]V)
}

// CompareAndSwap executes the compare-and-swap operation for a key.
// The RWMutexMap must have been initialized with an equal function, lest this function panics.
func (m *RWMutexMap[K, V]) CompareAndSwap(key K, oldValue, newValue V) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	current, exists := m.values[key]
	if !exists {
		// Handle case where key doesn't exist
		return false
	}

	if m.equal != nil {
		if m.equal(current, oldValue) {
			m.values[key] = newValue
			return true
		}
		return false
	}

	panic("called CompareAndSwap without equal function")
}

// Swap swaps the value for a key and returns the previous value if any.
func (m *RWMutexMap[K, V]) Swap(key K, value V) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldValue, loaded := m.values[key]
	m.values[key] = value
	if !loaded {
		var zero V
		return zero, false
	}
	return oldValue, true
}

// LoadOrStore returns the existing value for the key if present. Otherwise, it stores and returns
// the given value. The loaded result is true if the value was loaded, false if stored.
func (m *RWMutexMap[K, V]) LoadOrStore(key K, value V) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if v, ok := m.values[key]; ok {
		return v, true
	}
	m.values[key] = value
	return value, false
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
func (m *RWMutexMap[K, V]) LoadAndDelete(key K) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.values[key]
	if ok {
		delete(m.values, key)
		return v, true
	}
	var zero V
	return zero, false
}

// GetAll returns a copy of all key-value pairs in the map.
func (m *RWMutexMap[K, V]) GetAll() map[K]V {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[K]V)
	maps.Copy(result, m.values)
	return result
}

// GetMany retrieves multiple keys at once.
func (m *RWMutexMap[K, V]) GetMany(keys []K) map[K]V {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[K]V)
	for _, key := range keys {
		value, exists := m.values[key]
		if exists {
			result[key] = value
		}
	}
	return result
}

// SetMany sets multiple key-value pairs at once.
func (m *RWMutexMap[K, V]) SetMany(entries map[K]V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	maps.Insert(m.values, maps.All(entries))
}

// Equals reports whether the logical content of this map and the other map is the same. Requires
// equalFn to be provided to decide how two values of type V are compared.
func (m *RWMutexMap[K, V]) Equals(other Map[K, V], equalFn func(a, b V) bool) bool {
	return equals(m, other, equalFn)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m *RWMutexMap[K, V]) Range(f func(key K, value V) bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, v := range m.values {
		if !f(k, v) {
			break
		}
	}
}

// All returns an iterator over key-value pairs in the map.
// The iteration order is not guaranteed to be consistent.
func (m *RWMutexMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.mu.RLock()
		snapshot := maps.Clone(m.values)
		m.mu.RUnlock()

		for k, v := range snapshot {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Keys returns an iterator over keys in the map.
// The iteration order is not guaranteed to be consistent.
func (m *RWMutexMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		m.mu.RLock()
		keys := make([]K, 0, len(m.values))
		for k := range m.values {
			keys = append(keys, k)
		}
		m.mu.RUnlock()

		for _, k := range keys {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iterator over values in the map.
// The iteration order is not guaranteed to be consistent.
func (m *RWMutexMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		m.mu.RLock()
		values := make([]V, 0, len(m.values))
		for _, v := range m.values {
			values = append(values, v)
		}
		m.mu.RUnlock()

		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

// NewRWMutexMap creates a new instance of RWMutexMap.
func NewRWMutexMap[K comparable, V any](equalFn func(V, V) bool) *RWMutexMap[K, V] {
	return &RWMutexMap[K, V]{
		equal:  equalFn,
		values: make(map[K]V),
	}
}

// RWMutexMapFromMap creates a new instance of RWMutexMap from values in the provided map.
func RWMutexMapFromMap[K comparable, V any](m map[K]V, equalFn func(V, V) bool) *RWMutexMap[K, V] {
	newMap := NewRWMutexMap[K, V](equalFn)
	newMap.SetMany(m)
	return newMap
}
