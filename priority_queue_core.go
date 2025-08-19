// Package threadsafe implements thread-safe operations.
package threadsafe

import "sync"

// CorePriorityQueue is a minimal thread-safe priority queue that implements only
// the core PriorityQueue interface: Push, Pop, Peek, Len, Clear, Range.
// It does not expose any indexed mutation helpers, nor onSwap callbacks.
//
// It is a generic min-heap parameterized by a less comparator.
// The zero value is not ready; construct via NewCorePriorityQueue.
//
// Complexity: Push/Pop O(log n), Peek O(1); Range does not mutate the heap.
type CorePriorityQueue[T any] struct {
	mu    sync.RWMutex
	items []T
	less  func(a, b T) bool
}

// NewCorePriorityQueue creates a new minimal priority queue using the given comparator.
func NewCorePriorityQueue[T any](less func(a, b T) bool) *CorePriorityQueue[T] {
	return &CorePriorityQueue[T]{less: less}
}

// Push inserts one or more items into the queue.
func (q *CorePriorityQueue[T]) Push(items ...T) {
	if len(items) == 0 {
		return
	}
	q.mu.Lock()
	for _, x := range items {
		q.items = append(q.items, x)
		q.up(len(q.items) - 1)
	}
	q.mu.Unlock()
}

// Pop removes and returns the minimum item per the comparator.
func (q *CorePriorityQueue[T]) Pop() (item T, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return item, false
	}
	last := len(q.items) - 1
	q.swap(0, last)
	item = q.items[last]
	q.items = q.items[:last]
	if len(q.items) > 0 {
		q.down(0)
	}
	return item, true
}

// Peek returns the minimum item without removing it.
func (q *CorePriorityQueue[T]) Peek() (item T, ok bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if len(q.items) == 0 {
		return item, false
	}
	return q.items[0], true
}

// Len returns the number of items.
func (q *CorePriorityQueue[T]) Len() int {
	q.mu.RLock()
	l := len(q.items)
	q.mu.RUnlock()
	return l
}

// Clear removes all items.
func (q *CorePriorityQueue[T]) Clear() {
	q.mu.Lock()
	q.items = nil
	q.mu.Unlock()
}

// Range iterates over a snapshot of items in arbitrary internal order.
func (q *CorePriorityQueue[T]) Range(f func(item T) bool) {
	q.mu.RLock()
	snap := make([]T, len(q.items))
	copy(snap, q.items)
	q.mu.RUnlock()
	for _, it := range snap {
		if !f(it) {
			break
		}
	}
}

// Internal helpers (write-locked callers)
func (q *CorePriorityQueue[T]) lessIdx(i, j int) bool { return q.less(q.items[i], q.items[j]) }

func (q *CorePriorityQueue[T]) swap(i, j int) {
	if i == j {
		return
	}
	q.items[i], q.items[j] = q.items[j], q.items[i]
}

func (q *CorePriorityQueue[T]) up(i int) {
	for {
		p := (i - 1) / 2
		if i == 0 || !q.lessIdx(i, p) {
			break
		}
		q.swap(i, p)
		i = p
	}
}

// down moves item at i down; returns true if moved down.
func (q *CorePriorityQueue[T]) down(i int) bool {
	n := len(q.items)
	moved := false
	for {
		l := 2*i + 1
		if l >= n {
			break
		}
		smallest := l
		r := l + 1
		if r < n && q.lessIdx(r, l) {
			smallest = r
		}
		if !q.lessIdx(smallest, i) {
			break
		}
		q.swap(i, smallest)
		i = smallest
		moved = true
	}
	return moved
}

// Ensure CorePriorityQueue implements the core PriorityQueue interface.
var _ PriorityQueue[any] = (*CorePriorityQueue[any])(nil)
