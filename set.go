// Package threadsafe implements thread-safe operations.
package threadsafe

type Set[T any] interface {
    // Add stores an item in the set.
    Add(item T)
    // Remove deletes an item from the set.
    Remove(item T)
    // Has returns true if the item is in the set, otherwise false.
    Has(item T) bool
    // Len returns the number of items in the set.
    Len() int
    // Clear removes all items from the set.
    Clear()
}

