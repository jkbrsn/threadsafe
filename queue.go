// Package threadsafe implements thread-safe operations.
package threadsafe

// Queue is a generic FIFO queue interface for any type T.
// All operations must be safe for concurrent use by multiple goroutines.
//
// The contract is intentionally similar in style to Set and Map interfaces in this
// repository to provide a consistent developer experience.
type Queue[T any] interface {
	// Enqueue adds one or more items to the back of the queue.
	Enqueue(items ...T)

	// Pop removes and returns the item at the front of the queue.
	// If the queue is empty, it returns ok == false and the zero value of T.
	Pop() (item T, ok bool)

	// Peek returns the item at the front of the queue without removing it.
	// If the queue is empty, it returns ok == false and the zero value of T.
	Peek() (item T, ok bool)

	// Len returns the current number of items stored in the queue.
	Len() int

	// Clear removes all items from the queue.
	Clear()

	// Slice returns a copy of the current queue contents from front to back.
	// The returned slice is safe to read but may be stale if new items are added
	// concurrently.
	Slice() []T

	// Range calls f sequentially for each item present in the queue from front
	// to back. If f returns false, Range stops the iteration early.
	Range(f func(item T) bool)
}
