package threadsafe

import (
	"math/rand"
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// heapTestSuite is a generic test suite for the Heap interface.
// It is parameterized with a constructor and some sample items.
type heapTestSuite[T any] struct {
	newHeap func() Heap[T]
	less    func(a, b T) bool
	item1   T
	item2   T
	item3   T
}

func TestRWMutexHeapImplementsHeap(_ *testing.T) {
	// Compile-time assertion
	var _ Heap[int] = &RWMutexHeap[int]{}
}

// TestBasicOperations verifies Push, Pop, Peek, Len, Clear with ordering.
func (s *heapTestSuite[T]) TestBasicOperations(t *testing.T) {
	h := s.newHeap()
	assert.Equal(t, 0, h.Len())

	// Push items
	h.Push(s.item2, s.item1, s.item3)
	assert.Equal(t, 3, h.Len())

	// Peek should return top-priority item without removal
	item, ok := h.Peek()
	assert.True(t, ok)

	// Determine expected min based on provided less
	expectedTop := s.item1
	if s.less(s.item2, expectedTop) {
		expectedTop = s.item2
	}
	if s.less(s.item3, expectedTop) {
		expectedTop = s.item3
	}
	assert.Equal(t, expectedTop, item)
	assert.Equal(t, 3, h.Len())

	// Pop should return items in nondecreasing priority order according to less
	popped := make([]T, 0, 3)
	for i := 0; i < 3; i++ {
		it, ok := h.Pop()
		assert.True(t, ok)
		popped = append(popped, it)
	}
	assert.Equal(t, 0, h.Len())

	// Verify ordering using the same comparator
	isSorted := func(xs []T) bool {
		for i := 1; i < len(xs); i++ {
			if s.less(xs[i], xs[i-1]) { // out of order
				return false
			}
		}
		return true
	}
	assert.True(t, isSorted(popped))

	// Pop from empty
	_, ok = h.Pop()
	assert.False(t, ok)

	// Clear should be idempotent
	h.Clear()
	assert.Equal(t, 0, h.Len())
}

func (s *heapTestSuite[T]) TestSliceAndRange(t *testing.T) {
	h := s.newHeap()

	// Empty slice
	assert.Empty(t, h.Slice())

	// Push items and assert Slice returns copy (order not guaranteed sorted)
	h.Push(s.item1, s.item2, s.item3)
	sl := h.Slice()
	assert.Equal(t, 3, len(sl))

	// Range visits all items
	visited := []T{}
	h.Range(func(it T) bool {
		visited = append(visited, it)
		return true
	})
	assert.Equal(t, 3, len(visited))

	// Early stop
	count := 0
	h.Range(func(_ T) bool {
		count++
		return false
	})
	assert.Equal(t, 1, count)
}

func runHeapTestSuite[T any](t *testing.T, s *heapTestSuite[T]) {
	t.Run("BasicOperations", s.TestBasicOperations)
	t.Run("SliceAndRange", s.TestSliceAndRange)
}

func TestHeapImplementations(t *testing.T) {
	// int min-heap
	t.Run("int_min_heap", func(t *testing.T) {
		less := func(a, b int) bool { return a < b }
		makeHeap := func() Heap[int] { return NewRWMutexHeap(less) }
		suite := &heapTestSuite[int]{
			newHeap: makeHeap,
			less:    less,
			item1:   1,
			item2:   2,
			item3:   3,
		}
		runHeapTestSuite(t, suite)
	})

	// string min-heap
	t.Run("string_min_heap", func(t *testing.T) {
		less := func(a, b string) bool { return a < b }
		makeHeap := func() Heap[string] { return NewRWMutexHeap(less) }
		suite := &heapTestSuite[string]{
			newHeap: makeHeap,
			less:    less,
			item1:   "a",
			item2:   "b",
			item3:   "c",
		}
		runHeapTestSuite(t, suite)
	})

	// custom struct with priority
	t.Run("struct_custom_less", func(t *testing.T) {
		type task struct {
			ID       int
			Priority int
		}
		less := func(a, b task) bool { return a.Priority < b.Priority }
		makeHeap := func() Heap[task] { return NewRWMutexHeap(less) }
		suite := &heapTestSuite[task]{
			newHeap: makeHeap,
			less:    less,
			item1:   task{ID: 1, Priority: 10},
			item2:   task{ID: 2, Priority: 5},
			item3:   task{ID: 3, Priority: 7},
		}
		runHeapTestSuite(t, suite)
	})
}

// TestHeapPopOrder verifies that popping all elements yields a sorted sequence (by less).
func TestHeapPopOrder(t *testing.T) {
	less := func(a, b int) bool { return a < b }
	h := NewRWMutexHeap(less)
	// Insert random numbers
	rnd := rand.New(rand.NewSource(42))
	nums := make([]int, 200)
	for i := range nums {
		nums[i] = rnd.Intn(1000)
	}
	h.Push(nums...)

	// Pop all and verify sorted ascending
	out := make([]int, 0, len(nums))
	for {
		v, ok := h.Pop()
		if !ok {
			break
		}
		out = append(out, v)
	}
	assert.True(t, sort.IntsAreSorted(out))
}

// TestHeapConcurrentPush ensures thread-safety under concurrent pushes.
func TestHeapConcurrentPush(t *testing.T) {
	less := func(a, b int) bool { return a < b }
	h := NewRWMutexHeap(less)

	const goroutines = 8
	const perG = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(base int) {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				h.Push(base*perG + i)
			}
		}(g)
	}
	wg.Wait()

	// Pop all sequentially and ensure count matches
	total := goroutines * perG
	prev, ok := h.Pop()
	if ok {
		count := 1
		for {
			v, ok := h.Pop()
			if !ok {
				break
			}
			// non-decreasing order because it is a min-heap
			assert.LessOrEqual(t, prev, v)
			prev = v
			count++
		}
		assert.Equal(t, total, count)
	} else {
		assert.Equal(t, 0, total)
	}

	assert.Equal(t, 0, h.Len())
}
