// Package threadsafe implements thread-safe operations.
package threadsafe

// PriorityQueue is a generic thread-safe priority queue interface (min-heap) for any type T.
// Ordering is defined by the implementation at construction time via a comparator.
// All operations must be safe for concurrent use by multiple goroutines.
//
// Semantics mirror container/heap where applicable; indices are stable only for the
// lifetime between operations that may move elements. For external index maintenance
// (e.g., storing an "index" field inside elements), implementations may provide a
// swap-callback to notify callers when indices change.
//
// Complexity targets (amortized): Push/Pop O(log n), Peek O(1), Fix/RemoveAt O(log n).
// Snapshot/Range do not mutate the queue.
//
// Notes on mutability:
//   - If callers mutate ordering fields of an element already in the queue, they MUST
//     call Fix or UpdateAt to restore queue invariants.
//   - Peek returns a copy of T (or the pointer value for pointer T). Callers must
//     avoid in-place mutations without Fix/UpdateAt.
//
// Optional key support:
// Implementations may provide constructors that maintain an internal id->index map
// to enable O(log n) RemoveByKey/UpdateByKey helpers.
// These helpers are not part of the core interface to keep it generic.
//
// The interface is inspired by Queue/Set/Map in this repository for consistency.
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

	// Slice returns a copy of the queue's items in arbitrary internal order
	// (NOT sorted). This is intended for debugging or snapshotting.
	// If a sorted slice is needed, callers should copy and sort externally.
	Slice() []T

	// Range iterates over items in arbitrary internal order. Returning false stops early.
	Range(f func(item T) bool)

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
