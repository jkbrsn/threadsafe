// Package threadsafe implements thread-safe operations.
package threadsafe

import "iter"

// Slice is a generic interface for stores with any type T.
type Slice[T any] interface {
	// Append appends an item to the buffer in a thread-safe way.
	Append(item ...T)
	// Flush atomically retrieves all items and clears the buffer.
	// Returns a slice with the previous contents.
	Flush() []T
	// Peek returns a copy of the current buffer contents without clearing.
	// The returned slice is safe to read but may be stale if new items are added concurrently.
	Peek() []T
	// Len returns the current number of items in the buffer.
	Len() int

	// All returns an iterator over all items in the slice in order.
	//
	// Example usage:
	//
	//	for item := range mySlice.All() {
	//	    fmt.Println(item)
	//	}
	All() iter.Seq[T]
}
