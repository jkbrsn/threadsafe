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

func TestHeapPriorityQueueImplementsInterface(_ *testing.T) {
	var _ PriorityQueue[int] = &HeapPriorityQueue[int]{}
}

func TestRWMutexPriorityQueueImplementsInterface(_ *testing.T) {
	var _ PriorityQueue[int] = &RWMutexPriorityQueue[int]{}
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
	all := pq.Slice()
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
	snap := pq.Slice()
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
	if pq.Len() > 0 {
		_, ok := pq.RemoveAt(0)
		assert.True(t, ok)
	}

	// UpdateAt some index if exists and Fix
	if pq.Len() >= 1 {
		idx := 0
		x, _ := pq.Peek() // get a current min and reinsert as min again
		_ = pq.UpdateAt(idx, x)
		pq.Fix(idx)
	}
}

// runPriorityQueueTestSuite runs common tests for a PriorityQueue implementation.
func runPriorityQueueTestSuite[T any](t *testing.T, s *priorityQueueTestSuite[T]) {
	t.Run("BasicOperations", s.TestBasicOperations)
	t.Run("FixUpdateRemove", s.TestFixUpdateRemove)
	t.Run("ConcurrentOperations", s.TestConcurrentOperations)
}

// TestPriorityQueueImplementations runs the test suite for both implementations.
func TestPriorityQueueImplementations(t *testing.T) {
	items := func() []heapTestItem {
		return []heapTestItem{{ID: "a", Prio: 3}, {ID: "b", Prio: 1}, {ID: "c", Prio: 2}}
	}

	t.Run("RWMutexPriorityQueue", func(t *testing.T) {
		s := &priorityQueueTestSuite[heapTestItem]{
			newPQ: func() PriorityQueue[heapTestItem] {
				return NewRWMutexPriorityQueue(lessItem, onSwapItem)
			},
			less:  lessItem,
			prio:  func(x heapTestItem) int { return x.Prio },
			items: items,
		}
		runPriorityQueueTestSuite(t, s)
	})

	t.Run("HeapPriorityQueue", func(t *testing.T) {
		s := &priorityQueueTestSuite[heapTestItem]{
			newPQ: func() PriorityQueue[heapTestItem] {
				return NewHeapPriorityQueue(lessItem, nil)
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
		for i := 0; i < b.N; i++ { // preload with N items so we can pop in loop
			pq.Push(i)
		}
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
		// Pre-fill with some elements
		for i := 0; i < 1000; i++ {
			pq.Push(i)
		}
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

func BenchmarkPriorityQueueImplementations(b *testing.B) {
	b.Run("RWMutexPriorityQueue", func(b *testing.B) {
		benchmarkPriorityQueue(b, func() PriorityQueue[int] {
			return NewRWMutexPriorityQueue(func(a, b int) bool { return a < b }, nil)
		})
	})

	b.Run("HeapPriorityQueue", func(b *testing.B) {
		benchmarkPriorityQueue(b, func() PriorityQueue[int] {
			return NewHeapPriorityQueue(func(a, b int) bool { return a < b }, nil)
		})
	})
}
