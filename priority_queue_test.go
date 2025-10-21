package threadsafe

import (
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// heapTestItem is a sample element for testing PQs.
type heapTestItem struct {
	ID   string
	Prio int
	Idx  int // optional external index maintenance example
}

// helper less: min-heap by Prio
func lessItem(a, b heapTestItem) bool { return a.Prio < b.Prio }

// onSwap maintains external indices
func onSwapItem(i, j int, items []heapTestItem) {
	items[i].Idx = i
	items[j].Idx = j
}

func TestCorePriorityQueueImplementsInterface(_ *testing.T) {
	var _ PriorityQueue[int] = &CorePriorityQueue[int]{}
}

func TestIndexedPriorityQueueImplementsInterface(_ *testing.T) {
	var _ PriorityQueue[int] = &IndexedPriorityQueue[int]{}
	var _ PriorityQueueIndexed[int] = &IndexedPriorityQueue[int]{}
}

// priorityQueueTestSuite defines a reusable test suite for PriorityQueue[T].
// newPQ constructs a fresh queue for each test.
type priorityQueueTestSuite[T any] struct {
	// prio extracts a comparable priority for assertions across implementations
	prio  func(x T) int
	newPQ func() PriorityQueue[T]
	less  func(a, b T) bool
	// generator to produce some items for tests (unsorted)
	items func() []T
}

// TestConcurrentOperations pushes N random ints concurrently and then pops in order.
func (s *priorityQueueTestSuite[T]) TestConcurrentOperations(t *testing.T) {
	// This test is specialized to int T; guard via type assertion
	newIntPQ, ok := any(s.newPQ).(func() PriorityQueue[int])
	if !ok {
		// Skip for non-int specializations
		t.Skip("concurrency test defined for int priority only")
		return
	}
	const goroutines = 8
	const per = 200
	pq := newIntPQ()
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(seed int64) {
			defer wg.Done()
			r := rand.New(rand.NewSource(seed))
			vals := make([]int, per)
			for i := range per {
				vals[i] = r.Intn(1000000)
			}
			pq.Push(vals...)
		}(time.Now().UnixNano() + int64(g))
	}
	wg.Wait()
	// Snapshot via Range instead of Slice
	var all []int
	pq.Range(func(x int) bool { all = append(all, x); return true })
	sort.Ints(all)
	for _, want := range all {
		got, ok := pq.Pop()
		assert.True(t, ok)
		assert.Equal(t, want, got)
	}
	assert.Equal(t, 0, pq.Len())
}

func (s *priorityQueueTestSuite[T]) TestBasicOperations(t *testing.T) {
	pq := s.newPQ()
	assert.Equal(t, 0, pq.Len())

	itms := s.items()
	pq.Push(itms...)
	assert.Equal(t, len(itms), pq.Len())

	// Peek min then Pop in ascending order by comparator by comparing to sorted snapshot.
	var snap []T
	pq.Range(func(x T) bool { snap = append(snap, x); return true })
	// Make a sorted copy according to less
	sorted := make([]T, len(snap))
	copy(sorted, snap)
	sort.Slice(sorted, func(i, j int) bool { return s.less(sorted[i], sorted[j]) })

	first, ok := pq.Peek()
	assert.True(t, ok)
	// Compare by priority instead of full struct to ignore Idx changes
	assert.Equal(t, s.prio(sorted[0]), s.prio(first))

	for _, want := range sorted {
		got, ok := pq.Pop()
		assert.True(t, ok)
		assert.Equal(t, s.prio(want), s.prio(got))
	}
	_, ok = pq.Pop()
	assert.False(t, ok)

	// Clear after pushing again
	pq.Push(itms[0])
	assert.Equal(t, 1, pq.Len())
	pq.Clear()
	assert.Equal(t, 0, pq.Len())
}

func (s *priorityQueueTestSuite[T]) TestFixUpdateRemove(t *testing.T) {
	pq := s.newPQ()
	itms := s.items()
	pq.Push(itms...)

	// RemoveAt root if exists
	// If implementation supports indexed ops, exercise them
	if idxPQ, ok := any(pq).(PriorityQueueIndexed[T]); ok {
		if idxPQ.Len() > 0 {
			_, ok := idxPQ.RemoveAt(0)
			assert.True(t, ok)
		}
		if idxPQ.Len() >= 1 {
			idx := 0
			x, _ := idxPQ.Peek()
			_ = idxPQ.UpdateAt(idx, x)
			idxPQ.Fix(idx)
		}
	}
}

func (s *priorityQueueTestSuite[T]) TestAllIterator(t *testing.T) {
	pq := s.newPQ()
	itms := s.items()
	pq.Push(itms...)

	var snapshot []T
	pq.Range(func(x T) bool {
		snapshot = append(snapshot, x)
		return true
	})
	items := collectSeq(pq.All())
	assert.ElementsMatch(t, snapshot, items)

	var calls int
	pq.All()(func(_ T) bool {
		calls++
		return false
	})
	assert.Equal(t, 1, calls)

	var observed []T
	pq.All()(func(item T) bool {
		observed = append(observed, item)
		if len(observed) == 1 {
			pq.Push(s.items()[0])
		}
		return true
	})
	assert.ElementsMatch(t, snapshot, observed)
	assert.Equal(t, len(itms)+1, pq.Len())
}

// runPriorityQueueTestSuite runs common tests for a PriorityQueue implementation.
func runPriorityQueueTestSuite[T any](t *testing.T, s *priorityQueueTestSuite[T]) {
	t.Run("BasicOperations", s.TestBasicOperations)
	t.Run("FixUpdateRemove", s.TestFixUpdateRemove)
	t.Run("ConcurrentOperations", s.TestConcurrentOperations)
	t.Run("AllIterator", s.TestAllIterator)
}

// TestPriorityQueueImplementations runs the test suite for both implementations.
func TestPriorityQueueImplementations(t *testing.T) {
	items := func() []heapTestItem {
		return []heapTestItem{{ID: "a", Prio: 3}, {ID: "b", Prio: 1}, {ID: "c", Prio: 2}}
	}

	t.Run("CorePriorityQueue", func(t *testing.T) {
		s := &priorityQueueTestSuite[heapTestItem]{
			newPQ: func() PriorityQueue[heapTestItem] { return NewCorePriorityQueue(lessItem) },
			less:  lessItem,
			prio:  func(x heapTestItem) int { return x.Prio },
			items: items,
		}
		runPriorityQueueTestSuite(t, s)
	})

	t.Run("IndexedPriorityQueue", func(t *testing.T) {
		s := &priorityQueueTestSuite[heapTestItem]{
			newPQ: func() PriorityQueue[heapTestItem] {
				return NewIndexedPriorityQueue(lessItem, onSwapItem)
			},
			less:  lessItem,
			prio:  func(x heapTestItem) int { return x.Prio },
			items: items,
		}
		runPriorityQueueTestSuite(t, s)
	})
}

//
// BENCHMARKS
//

// benchmarkPriorityQueue exercises common PQ operations.
func benchmarkPriorityQueue(b *testing.B, newPQ func() PriorityQueue[int]) {
	// Push benchmark
	b.Run("Push", func(b *testing.B) {
		pq := newPQ()
		b.ResetTimer()
		for b.Loop() {
			pq.Push(1)
		}
	})

	// Peek benchmark
	b.Run("Peek", func(b *testing.B) {
		pq := newPQ()
		pq.Push(1)
		b.ResetTimer()
		for b.Loop() {
			pq.Peek()
		}
	})

	// Pop benchmark
	b.Run("Pop", func(b *testing.B) {
		pq := newPQ()
		// Preload with N items
		fillPQ(pq, b.N)
		b.ResetTimer()
		for b.Loop() {
			if _, ok := pq.Pop(); !ok {
				// Refill minimally to keep popping
				pq.Push(1)
			}
		}
	})

	// Mixed concurrent workload (approx 80% Peek, 15% Push, 5% Pop)
	b.Run("ConcurrentMixed", func(b *testing.B) {
		pq := newPQ()
		// Pre-fill with 1000 items
		fillPQ(pq, 1000)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				i++
				s := i % 20
				if s == 0 { // 5%
					pq.Pop()
				} else if s%4 == 0 { // 15%
					pq.Push(i)
				} else { // 80%
					pq.Peek()
				}
			}
		})
	})
}

func fillPQ(pq PriorityQueue[int], n int) {
	for i := range n {
		pq.Push(i)
	}
}

func BenchmarkPriorityQueueImplementations(b *testing.B) {
	b.Run("CorePriorityQueue", func(b *testing.B) {
		benchmarkPriorityQueue(b, func() PriorityQueue[int] {
			return NewCorePriorityQueue(func(a, b int) bool { return a < b })
		})
	})

	b.Run("IndexedPriorityQueue", func(b *testing.B) {
		benchmarkPriorityQueue(b, func() PriorityQueue[int] {
			return NewIndexedPriorityQueue(func(a, b int) bool { return a < b }, nil)
		})
	})
}
