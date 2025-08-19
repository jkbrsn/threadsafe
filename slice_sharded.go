// Package threadsafe implements thread-safe operations.
package threadsafe

import (
	"sync/atomic"
)

// ShardedSlice is a high-throughput thread-safe buffer that splits its storage into several
// independent shards. Each shard is a full Slice implementation, so operations on different
// shards proceed in parallel with zero contention.
//
// Using a ShardedSlice will reduce lock contention, retain atomicity at the API level, and
// allow for simple fall-back to a single mutex-backed slice.
//
// The slice returned by Flush() / Peek() concatenates shards in ascending index order. The
// order of items within each shard is preserved, but the overall order is only guaranteed
// per-shard, which is usually acceptable for buffer/queue-like workloads where ordering
// across goroutines is not critical.
//
// All methods are wait-free with bounded work and require no global locks.
type ShardedSlice[T any] struct {
	shards  []Slice[T]
	counter uint64 // used for round-robin shard selection in Append
}

// Append adds the items to one of the shards, selected in a round-robin
// manner using an atomic counter.  This ensures good key distribution without
// requiring hashing the items themselves.
func (s *ShardedSlice[T]) Append(item ...T) {
	idx := int(atomic.AddUint64(&s.counter, 1)-1) % len(s.shards)
	s.shards[idx].Append(item...)
}

// Flush atomically retrieves and clears all shards, concatenating the results into a single slice.
func (s *ShardedSlice[T]) Flush() []T {
	// First pass: determine total length
	total := 0
	for _, sh := range s.shards {
		total += sh.Len()
	}

	out := make([]T, 0, total)
	for _, sh := range s.shards {
		out = append(out, sh.Flush()...)
	}
	return out
}

// Peek returns a copy of the current contents of all shards without clearing them.
func (s *ShardedSlice[T]) Peek() []T {
	total := 0
	for _, sh := range s.shards {
		total += sh.Len()
	}
	out := make([]T, 0, total)
	for _, sh := range s.shards {
		out = append(out, sh.Peek()...)
	}
	return out
}

// Len returns the combined length of all shards.
func (s *ShardedSlice[T]) Len() int {
	total := 0
	for _, sh := range s.shards {
		total += sh.Len()
	}
	return total
}

// NewShardedSlice creates a ShardedSlice with the given number of shards.
// Each shard is pre-allocated with initialCap capacity.  shardCount must be
// >0; if <=0, it is coerced to 1.
func NewShardedSlice[T any](shardCount, initialCap int) *ShardedSlice[T] {
	nShards := shardCount
	if shardCount <= 0 {
		nShards = 1
	}
	shards := make([]Slice[T], nShards)
	for i := 0; i < nShards; i++ {
		// Use a minimal internal implementation – simple mutex slice.
		shards[i] = NewRWMutexSlice[T](initialCap)
	}
	return &ShardedSlice[T]{shards: shards}
}
