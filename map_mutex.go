// Package threadsafe implements thread-safe operations.
package threadsafe

import (
	"maps"
	"sync"
)

// MutexMap is a thread-safe implementation of Map using sync.Mutex.
type MutexMap[K comparable, V any] struct {
	mu     sync.Mutex
	values map[K]V

	equal func(V, V) bool
}

// Get retrieves the value for the given key.
func (m *MutexMap[K, V]) Get(key K) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	value, ok := m.values[key]
	return value, ok
}

// Set stores a value for the given key.
func (m *MutexMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.values[key] = value
}

// Delete removes the key from the map.
func (m *MutexMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.values, key)
}

// Len returns the number of items in the map.
func (m *MutexMap[K, V]) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.values)
}

// Clear removes all items from the map.
func (m *MutexMap[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.values = make(map[K]V)
}

// CompareAndSwap executes the compare-and-swap operation for a key.
// The MutexMap must have been initialized with an equal function, lest this function panics.
func (m *MutexMap[K, V]) CompareAndSwap(key K, oldValue, newValue V) bool {
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
func (m *MutexMap[K, V]) Swap(key K, value V) (V, bool) {
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

// GetAll returns a copy of all key-value pairs in the map.
func (m *MutexMap[K, V]) GetAll() map[K]V {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(map[K]V)
	maps.Copy(result, m.values)
	return result
}

// GetMany retrieves multiple keys at once.
func (m *MutexMap[K, V]) GetMany(keys []K) map[K]V {
	m.mu.Lock()
	defer m.mu.Unlock()

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
func (m *MutexMap[K, V]) SetMany(entries map[K]V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	maps.Copy(m.values, entries)
}

// Equals reports whether the logical content of this map and the other map is the same. Requires
// equalFn to be provided to decide how two values of type V are compared.
func (m *MutexMap[K, V]) Equals(other Map[K, V], equalFn func(a, b V) bool) bool {
	return equals(m, other, equalFn)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m *MutexMap[K, V]) Range(f func(key K, value V) bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, v := range m.values {
		if !f(k, v) {
			break
		}
	}
}

// NewMutexMap creates a new instance of MutexMap.
func NewMutexMap[K comparable, V any](equalFn func(V, V) bool) *MutexMap[K, V] {
	return &MutexMap[K, V]{
		equal:  equalFn,
		values: make(map[K]V),
	}
}
