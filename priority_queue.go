// Package threadsafe implements thread-safe operations.
package threadsafe

import "iter"

// PriorityQueue is a generic thread-safe priority queue interface (min-heap) for any type T.
// Ordering is defined by the implementation at construction time via a comparator. Implementations
// of this interface are expected to be safe for parallel use in multiple goroutines.
type PriorityQueue[T any] interface {
	// Push inserts one or more items into the queue.
	Push(items ...T)

	// Pop removes and returns the minimum item per the comparator.
	// If empty, returns ok == false and the zero value of T.
	Pop() (item T, ok bool)

	// Peek returns the current minimum without removing it.
	// If empty, returns ok == false and the zero value of T.
	Peek() (item T, ok bool)

	// Len returns the number of items in the queue.
	Len() int

	// Clear removes all items from the queue.
	Clear()

	// Range iterates over items in arbitrary internal order. Returning false stops early.
	Range(f func(item T) bool)

	// All returns an iterator over items in the queue in internal heap order (not sorted).
	// The iteration order is implementation-defined and not guaranteed to be priority-sorted.
	//
	// Example usage:
	//
	//	for item := range myPQ.All() {
	//	    fmt.Println(item)
	//	}
	All() iter.Seq[T]
}

// PriorityQueueIndexed exposes index-based mutation helpers intended for advanced use-cases.
//
// As the index-based helpers brings on mutation risks it's important to note:
//   - If callers mutate ordering fields of an element already in the queue, they MUST
//     call Fix or UpdateAt to restore queue invariants.
//   - Peek returns a copy of T (or the pointer value for pointer T). Callers must
//     avoid in-place mutations without Fix/UpdateAt.
type PriorityQueueIndexed[T any] interface {
	PriorityQueue[T]

	// Fix re-establishes queue ordering after the item at index i may have changed.
	// Safe no-op if i is out of range.
	Fix(i int)

	// RemoveAt removes and returns the item at index i in the queue's internal array.
	// If i is out of range, returns zero value and ok == false.
	RemoveAt(i int) (item T, ok bool)

	// UpdateAt replaces the element at index i with x and restores queue invariants.
	// If i is out of range, it is a no-op and returns false.
	UpdateAt(i int, x T) bool
}
