// Package threadsafe implements thread-safe operations.
package threadsafe

// Map is a generic interface for stores with any type V.
// It allows concurrent appends and atomic flushes.
type Map[K comparable, V any] interface {
	// Basic operations
	Get(key K) (V, bool) // Returns value and existence flag
	Set(key K, value V)
	Delete(key K)
	Len() int
	Clear()

	// Conditional operations
	CompareAndSwap(key K, oldValue, newValue V) bool
	Swap(key K, value V) (previous V, loaded bool)

	// Batch operations
	GetAll() map[K]V
	GetMany(keys []K) map[K]V
	SetMany(entries map[K]V)

	// Iteration
	Range(func(key K, value V) bool)
}

// MapDiff represents the difference between two maps.
type MapDiff[K comparable, V any] struct {
	AddedOrModified map[K]V
	Removed         map[K]V
}

// CalculateMapDiff calculates the difference between two maps.
// It returns a MapDiff containing the added or modified entries and the removed entries.
func CalculateMapDiff[K comparable, V any](
	newData, oldData map[K]V,
	equalFunc func(V, V) bool,
) MapDiff[K, V] {
	diff := MapDiff[K, V]{
		AddedOrModified: make(map[K]V),
		Removed:         make(map[K]V),
	}

	// Check for new or modified entries
	for k, newVal := range newData {
		if oldVal, exists := oldData[k]; !exists || !equalFunc(oldVal, newVal) {
			diff.AddedOrModified[k] = newVal
		}
	}

	// Check for removed entries
	for k := range oldData {
		if _, exists := newData[k]; !exists {
			var zero V
			diff.Removed[k] = zero // or store the old value if needed
		}
	}

	return diff
}
