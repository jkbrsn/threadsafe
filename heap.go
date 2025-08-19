// Package threadsafe implements thread-safe operations.
package threadsafe

// Heap is a generic binary heap interface for any type T.
// Ordering is defined by the implementation, typically via a provided less function.
// All operations are expected to be safe for concurrent use by multiple goroutines.
type Heap[T any] interface {
	// Push adds one or more items to the heap.
	Push(items ...T)

	// Pop removes and returns the top-priority item.
	// If the heap is empty, it returns ok == false and the zero value of T.
	Pop() (item T, ok bool)

	// Peek returns the top-priority item without removing it.
	// If the heap is empty, it returns ok == false and the zero value of T.
	Peek() (item T, ok bool)

	// Len returns the current number of items stored in the heap.
	Len() int

	// Clear removes all items from the heap.
	Clear()

	// Slice returns a copy of the current heap contents in internal heap order
	// (not sorted). The returned slice is safe to read but may be stale if
	// new items are added concurrently.
	Slice() []T

	// Range calls f sequentially for each item present in the heap in internal
	// heap order. If f returns false, Range stops the iteration early.
	Range(f func(item T) bool)
}
