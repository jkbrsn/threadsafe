package threadsafe

import (
    "sort"
    "strconv"
    "sync"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestSyncMapSetImplementsSet(_ *testing.T) {
    var _ Set[string] = &SyncMapSet[string]{}
}

func TestSyncMapSetBasicOperations(t *testing.T) {
    set := NewSyncMapSet[string]()

    // Initial state
    assert.Equal(t, 0, set.Len())
    assert.False(t, set.Has("item1"))

    // Add
    set.Add("item1")
    assert.Equal(t, 1, set.Len())
    assert.True(t, set.Has("item1"))

    // Add duplicate
    set.Add("item1")
    assert.Equal(t, 1, set.Len())

    // Add multiple
    set.Add("item2")
    set.Add("item3")
    assert.Equal(t, 3, set.Len())

    // Remove
    set.Remove("item2")
    assert.Equal(t, 2, set.Len())
    assert.False(t, set.Has("item2"))

    // Clear
    set.Clear()
    assert.Equal(t, 0, set.Len())
}

func TestSyncMapSetSlice(t *testing.T) {
    set := NewSyncMapSet[int]()
    items := []int{3, 1, 4, 1, 5, 9}
    for _, it := range items {
        set.Add(it)
    }
    s := set.Slice()
    sort.Ints(s)
    expected := []int{1, 3, 4, 5, 9}
    assert.Equal(t, expected, s)
}

func TestSyncMapSetRange(t *testing.T) {
    set := NewSyncMapSet[string]()
    items := []string{"apple", "banana", "cherry"}
    for _, it := range items {
        set.Add(it)
    }
    visited := []string{}
    set.Range(func(item string) bool {
        visited = append(visited, item)
        return true
    })
    sort.Strings(visited)
    sort.Strings(items)
    assert.Equal(t, items, visited)

    // early stop
    count := 0
    set.Range(func(item string) bool {
        count++
        return count < 2
    })
    assert.Equal(t, 2, count)
}

func TestSyncMapSetConcurrentAccess(t *testing.T) {
    set := NewSyncMapSet[int]()
    const g = 100
    const ops = 100
    var wg sync.WaitGroup

    for i := 0; i < g; i++ {
        wg.Add(1)
        go func(base int) {
            defer wg.Done()
            for j := 0; j < ops; j++ {
                set.Add(base*ops + j)
            }
        }(i)
    }

    wg.Wait()
    assert.Equal(t, g*ops, set.Len())
}

func TestSyncMapSetConcurrentRemoval(t *testing.T) {
    set := NewSyncMapSet[string]()
    const n = 1000
    for i := 0; i < n; i++ {
        set.Add("item" + strconv.Itoa(i))
    }
    var wg sync.WaitGroup
    for i := 0; i < n; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            set.Remove("item" + strconv.Itoa(idx))
        }(i)
    }
    wg.Wait()
    assert.Equal(t, 0, set.Len())
}

func BenchmarkSyncMapSetAdd(b *testing.B) {
    set := NewSyncMapSet[int]()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        set.Add(i)
    }
}

func BenchmarkSyncMapSetHas(b *testing.B) {
    set := NewSyncMapSet[int]()
    for i := 0; i < 1000; i++ {
        set.Add(i)
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        set.Has(i % 1000)
    }
}

func BenchmarkSyncMapSetConcurrentReadWrite(b *testing.B) {
    set := NewSyncMapSet[int]()
    for i := 0; i < 1000; i++ {
        set.Add(i)
    }
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            if i%2 == 0 {
                set.Has(i % 1000)
            } else {
                set.Add(i + 1000)
            }
            i++
        }
    })
}
