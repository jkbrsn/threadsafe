// Package threadsafe implements thread-safe operations.
package threadsafe

import "sync"

// CorePriorityQueue is a thread-safe priority queue that implements the core PriorityQueue
// interface. It does not expose any indexed mutation helpers, nor onSwap callbacks.
//
// It is a generic min-heap parameterized by a less comparator. The zero value is not ready;
// construct via NewCorePriorityQueue. The less(a,b) comparator must define a strict weak ordering
// (irreflexive, transitive, consistent).
//
// Complexity: Push/Pop O(log n), Peek O(1); Range does not mutate the heap.
type CorePriorityQueue[T any] struct {
	mu    sync.RWMutex
	items []T
	less  func(a, b T) bool
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

// Range iterates over a snapshot of items in arbitrary internal order. Mutations during range
// does not affect the current iteration.
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
	idx := i
	for {
		p := (idx - 1) / 2
		if idx == 0 || !q.lessIdx(idx, p) {
			break
		}
		q.swap(idx, p)
		idx = p
	}
}

// down moves item at i down; returns true if moved down.
func (q *CorePriorityQueue[T]) down(i int) bool {
	idx := i
	n := len(q.items)
	moved := false
	for {
		l := 2*idx + 1
		if l >= n {
			break
		}
		smallest := l
		r := l + 1
		if r < n && q.lessIdx(r, l) {
			smallest = r
		}
		if !q.lessIdx(smallest, idx) {
			break
		}
		q.swap(idx, smallest)
		idx = smallest
		moved = true
	}
	return moved
}

// NewCorePriorityQueue creates a new minimal priority queue using the given comparator.
func NewCorePriorityQueue[T any](less func(a, b T) bool) *CorePriorityQueue[T] {
	return &CorePriorityQueue[T]{less: less}
}
