// Package threadsafe implements thread-safe operations.
package threadsafe

// Map is a generic interface for stores with any type V.
// It allows concurrent appends and atomic flushes.
type Map[K comparable, V any] interface {
	// Get retrieves the value for the given key.
	Get(key K) (V, bool)
	// Set stores a value for the given key.
	Set(key K, value V)
	// Delete removes the key from the map.
	Delete(key K)
	// Len returns the number of items in the map.
	Len() int
	// Clear removes all items from the map.
	Clear()

	// CompareAndSwap executes the compare-and-swap operation for a key.
	CompareAndSwap(key K, oldValue, newValue V) bool
	// Swap swaps the value for a key and returns the previous value if any.
	Swap(key K, value V) (previous V, loaded bool)

	// GetAll returns all key-value pairs in the map.
	GetAll() map[K]V
	// GetMany retrieves select key-value pairs.
	GetMany(keys []K) map[K]V
	// SetMany sets multiple key-value pairs.
	SetMany(entries map[K]V)

	// Equals reports whether the logical content of this map and the other map is the same.
	// Requires an equal function since V is not of type comparable.
	Equals(other Map[K, V], equalFn func(a, b V) bool) bool

	// Range calls f sequentially for each key and value present in the map.
	// If f returns false, range stops the iteration.
	Range(f func(key K, value V) bool)
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

// equals reports whether the logical content of two maps is the same. The comparison method is
// based on the equalFn provided.
func equals[K comparable, V any](
	a, b Map[K, V],
	equalFn func(V, V) bool) bool {

	// Fast paths: check object pointers and lengths
	if &a == &b {
		return true
	}
	if a.Len() != b.Len() {
		return false
	}

	// Snapshot each map once to avoid races and keep O(n) complexity
	am := a.GetAll()
	bm := b.GetAll()

	// Compare key-value pairs
	for k, av := range am {
		bv, ok := bm[k]
		if !ok || !equalFn(av, bv) {
			return false
		}
	}
	return true
}
