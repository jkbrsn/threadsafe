// Package threadsafe implements thread-safe operations.
package threadsafe

// Set is a generic interface for a set store of any type T.
type Set[T comparable] interface {
	// Add stores an item in the set.
	Add(item T) (added bool)
	// Delete removes an item from the set.
	Delete(item T) (removed bool)
	// Has returns true if the item is in the set, otherwise false.
	Has(item T) bool
	// Len returns the number of items in the set.
	Len() int
	// Clear removes all items from the set.
	Clear()
	// Slice returns a copy of the set as a slice.
	Slice() []T
	// Range calls f sequentially for each item present in the set.
	// If f returns false, range stops the iteration.
	Range(f func(item T) bool)
}
