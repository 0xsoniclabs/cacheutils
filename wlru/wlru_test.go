package wlru

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_InvalidParameters(t *testing.T) {
	_, err := New(10, -10)
	assert.Error(t, err)
}

func TestAdd_EvictionAndWeightManagement(t *testing.T) {
	cache, _ := New(5, 5)

	cache.Add(1, 1, 1)
	cache.Add(2, 2, 2)              // Weight: 3
	cache.Add(2, 3, 2)              // Update existing key
	assert.Equal(t, 2, cache.Len()) // Keys: 1,2 (no eviction yet)
	assert.Equal(t, uint(3), cache.Weight())

	evicted := cache.Add(3, 3, 3)            // Total would be 6 - triggers eviction
	assert.Equal(t, 1, evicted)              // Evicted 1 item (key 1)
	assert.Equal(t, 2, cache.Len())          // Keys: 2,3
	assert.Equal(t, uint(5), cache.Weight()) // 2 + 3
}

func TestTotal_ReturnsAccurateMetrics(t *testing.T) {
	cache, _ := New(5, 5)
	cache.Add(1, 1, 1)
	cache.Add(2, 2, 2)

	weight, num := cache.Total()
	assert.Equal(t, uint(3), weight)
	assert.Equal(t, 2, num)
}

func TestGet_Operations(t *testing.T) {
	cache, _ := New(5, 5)
	cache.Add(2, 3, 2)

	val, ok := cache.Get(2)
	assert.True(t, ok)
	assert.Equal(t, 3, val)

	val, ok = cache.Get(99)
	assert.False(t, ok)
}

func TestContains_KeyVerification(t *testing.T) {
	cache, _ := New(5, 5)
	cache.Add(2, 3, 2)

	assert.True(t, cache.Contains(2))
	assert.False(t, cache.Contains(99))
}

func TestPeek_NonMutatingAccess(t *testing.T) {
	cache, _ := New(5, 5)
	cache.Add(1, 1, 1)
	cache.Add(2, 3, 2)

	val, _ := cache.Peek(2)
	assert.Equal(t, 3, val)

	// Verify order remains unchanged - Peek does not mutate order
	k, _, _ := cache.GetOldest()
	assert.Equal(t, 1, k)
}

func TestKeys_OrderAndCompleteness(t *testing.T) {
	cache, _ := New(5, 5)
	cache.Add(1, 1, 1)
	cache.Add(2, 2, 2)

	keys := cache.Keys()
	assert.ElementsMatch(t, []interface{}{1, 2}, keys)
}

func TestOldest_Operations(t *testing.T) {
	cache, _ := New(5, 5)
	cache.Add(1, 1, 1)
	cache.Add(2, 2, 2)

	k, v, ok := cache.GetOldest()
	assert.True(t, ok)
	assert.Equal(t, 1, k)
	assert.Equal(t, 1, v)

	k, v, ok = cache.RemoveOldest()
	assert.True(t, ok)
	assert.Equal(t, 1, k)
	assert.False(t, cache.Contains(1))
}

func TestContainsOrAdd_KeyManagement(t *testing.T) {
	cache, _ := New(5, 5)
	cache.Add(2, 3, 2)

	exists, evicted := cache.ContainsOrAdd(2, "new", 1)
	assert.True(t, exists)
	assert.Equal(t, 0, evicted)
}

func TestPeekOrAdd_Operations(t *testing.T) {
	cache, _ := New(3, 2)
	cache.Add(1, "A", 2)

	// Existing key
	val, exists, _ := cache.PeekOrAdd(1, "B", 1)
	assert.Equal(t, "A", val)
	assert.True(t, exists)

	// New key with eviction
	_, _, evicted := cache.PeekOrAdd(2, "C", 2)
	assert.Equal(t, 1, evicted)
}

func TestResize_AdjustsCacheParameters(t *testing.T) {
	cache, _ := New(5, 5)
	cache.Add(1, 1, 1)
	cache.Add(2, 2, 1)
	cache.Add(3, 3, 2)

	evicted := cache.Resize(3, 3)
	assert.Equal(t, 1, evicted)
	assert.Equal(t, 2, cache.Len())
}

func TestRemove_EntryDeletion(t *testing.T) {
	cache, _ := New(5, 5)
	cache.Add(1, 1, 1)

	cache.Remove(1)
	assert.False(t, cache.Contains(1))
}

func TestPurge_CacheReset(t *testing.T) {
	cache, _ := New(5, 5)
	cache.Add(1, 1, 1)

	cache.Purge()
	assert.Equal(t, 0, cache.Len())
	assert.Equal(t, uint(0), cache.Weight())
}

func TestContainsOrAdd_EvictsWhenNeeded(t *testing.T) {
	cache, _ := New(5, 5)
	cache.Add(1, "A", 3)
	cache.Add(2, "B", 2)

	exists, evicted := cache.ContainsOrAdd(3, "C", 3)
	assert.False(t, exists)
	assert.Equal(t, 1, evicted) // Evicted oldest item (weight 3)
}

func TestPeekOrAdd_EvictsForNewEntries(t *testing.T) {
	cache, _ := New(3, 2)
	cache.Add(1, "A", 2)
	cache.Add(2, "B", 1)

	_, _, evicted := cache.PeekOrAdd(3, "C", 1)
	assert.Equal(t, 1, evicted) // Evicted weight 2 entry
}
