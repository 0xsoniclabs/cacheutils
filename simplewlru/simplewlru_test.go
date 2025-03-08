package simplewlru

import (
	"testing"
)

func TestNew(t *testing.T) {
	if c, err := New(10, 3); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	} else if c == nil {
		t.Fatalf("expected a valid cache, got nil")
	}
}

func TestNewWithNegativeSize(t *testing.T) {
	c, err := NewWithEvict(10, -1, nil)
	if err == nil {
		t.Errorf("expected error for negative maxSize, got cache: %+v", c)
	}
}

func TestPurge(t *testing.T) {
	// Use a callback to count evictions.
	var count int
	onEvict := func(key, value interface{}) {
		count++
	}
	c, _ := NewWithEvict(100, 10, onEvict)
	c.Add("x", "X", 10)
	c.Add("y", "Y", 10)
	c.Add("z", "Z", 10)
	if c.Len() != 3 {
		t.Errorf("expected 3 items before purge, got %d", c.Len())
	}
	c.Purge()
	if c.Len() != 0 {
		t.Errorf("expected 0 items after purge, got %d", c.Len())
	}
	if count != 3 {
		t.Errorf("expected 3 evictions from purge, got %d", count)
	}
	if c.Weight() != 0 {
		t.Errorf("expected total weight 0 after purge, got %d", c.Weight())
	}
}

func TestPurgeEmptyCache(t *testing.T) {
	c, _ := New(100, 10)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("purge on empty cache panicked: %v", r)
		}
	}()
	c.Purge()
}

func TestAddAndGet(t *testing.T) {
	c, _ := New(100, 10)
	evicted := c.Add("a", "apple", 10)
	if evicted != 0 {
		t.Errorf("unexpected eviction on first add, got %d", evicted)
	}

	value, ok := c.Get("a")
	if !ok {
		t.Errorf("expected key 'a' to be found")
	}
	if value != "apple" {
		t.Errorf("expected value 'apple', got %v", value)
	}

	evicted = c.Add("a", "apricot", 15)
	if evicted != 0 {
		t.Errorf("unexpected eviction on update, got %d", evicted)
	}
	value, ok = c.Get("a")
	if !ok || value != "apricot" {
		t.Errorf("update failed: want %v, got %v", "apricot", value)
	}

	peek, ok := c.Peek("a")
	if !ok || peek != "apricot" {
		t.Errorf("peek failed: want %v, got %v", "apricot", peek)
	}
}

func TestUpdateItemWeight(t *testing.T) {
	c, _ := New(50, 10)
	c.Add("key", "value1", 20)
	c.Add("key", "value2", 5)
	if c.Weight() != 5 {
		t.Errorf("expected weight to be updated to 5, got %d", c.Weight())
	}
	val, ok := c.Get("key")
	if !ok || val != "value2" {
		t.Errorf("expected value 'value2', got %v", val)
	}
}

func TestMultipleEvictionsByWeight(t *testing.T) {
	c, _ := New(30, 5)
	c.Add("a", 1, 10)            // weight=10
	c.Add("b", 2, 10)            // weight=20
	c.Add("c", 3, 10)            // weight=30
	evicted := c.Add("d", 4, 20) // total would be 40 if none evicted, need to evict at least one.
	if c.Weight() > 30 {
		t.Errorf("expected weight <= 30 after eviction, got %d", c.Weight())
	}
	if evicted != 2 {
		t.Errorf("expected two evictions from weight constraint, got %d", evicted)
	}
	if c.Len() > 5 {
		t.Errorf("cache length exceeded maxSize: got %d", c.Len())
	}
	if c.Len() != 2 {
		t.Errorf("cache length expected to be 2: got %d", c.Len())
	}
}

func TestEvictionBySize(t *testing.T) {
	c, _ := New(50, 3)
	c.Add("a", 1, 10)            // weight=10, size=1
	c.Add("b", 2, 10)            // weight=20, size=2
	c.Add("c", 3, 10)            // weight=30, size=3
	evicted := c.Add("d", 4, 20) // size would now be 4 if none evicted, need to evict
	if c.Weight() > 40 {
		t.Errorf("expected weight <= 40 after eviction, got %d", c.Weight())
	}
	if evicted != 1 {
		t.Errorf("expected one eviction from size constraint, got %d", evicted)
	}
	if c.Len() > 3 {
		t.Errorf("cache length exceeded maxSize: got %d", c.Len())
	}
	if c.Len() != 3 {
		t.Errorf("cache length expected to be 3: got %d", c.Len())
	}
}

func TestErrorHandlingOnEvictCallback(t *testing.T) {
	onEvict := func(key, value interface{}) {
		if key == "panic" {
			panic("forced panic")
		}
	}
	c, _ := NewWithEvict(100, 1, onEvict)
	c.Add("panic", "fail", 10)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for key 'panic' but did not panic")
		} else if r != "forced panic" {
			t.Errorf("unexpected panic value: %v", r)
		}
	}()
	c.Add("keep", "ok", 10)
}

func TestContainsAndRemove(t *testing.T) {
	c, _ := New(50, 5)
	c.Add("x", 100, 5)
	if !c.Contains("x") {
		t.Errorf("contains failed: expected key 'x' to be present")
	}
	present := c.Remove("x")
	if !present {
		t.Errorf("remove failed: expected key 'x' to be removed")
	}
	if c.Contains("x") {
		t.Errorf("expected key 'x' to be absent after removal")
	}
	present = c.Remove("nonexistent")
	if present {
		t.Errorf("remove should return false for key that does not exist")
	}
}

func TestRemoveElement(t *testing.T) {
	c, _ := New(100, 10)
	c.Add("a", 1, 5)
	c.Add("b", 2, 5)
	if !c.Remove("a") {
		t.Errorf("expected Remove to succeed for key 'a'")
	}
	if c.Contains("a") {
		t.Errorf("internal map still has key 'a'")
	}
	if c.Weight() != 5 {
		t.Errorf("expected weight 5 after removal, got %d", c.Weight())
	}
}

func TestGetNonExistent(t *testing.T) {
	c, _ := New(100, 10)
	val, ok := c.Get("nonexistent")
	if ok {
		t.Errorf("expected key 'nonexistent' to be absent, got value %v", val)
	}

	val, ok = c.Peek("nonexistent")
	if ok {
		t.Errorf("expected peek on 'nonexistent' to return false, got value %v", val)
	}
}

func TestGetWithNilEntry(t *testing.T) {
	c, _ := New(100, 10)
	key := "nilEntryKey"

	// Manually add an element with a nil *entry to the cache
	element := c.evictList.PushFront((*entry)(nil))
	c.items[key] = element

	value, ok := c.Get(key)
	if ok {
		t.Error("expected ok to be false for nil entry")
	}
	if value != nil {
		t.Errorf("expected value to be nil, got %v", value)
	}
}

func TestRemoveOldestAndGetOldest(t *testing.T) {
	c, _ := New(100, 10)
	c.Add("first", 1, 1)
	c.Add("second", 2, 1)
	c.Add("third", 3, 1)

	key, val, ok := c.GetOldest()
	if !ok {
		t.Errorf("expected to get an oldest element")
	}
	if key != "first" || val != 1 {
		t.Errorf("expected oldest to be ('first', 1), got (%v, %v)", key, val)
	}

	remKey, remVal, ok := c.RemoveOldest()
	if !ok {
		t.Errorf("expected removal of oldest element")
	}
	if remKey != "first" || remVal != 1 {
		t.Errorf("expected removed oldest key to be ('first', 1), got (%v, %v)", remKey, remVal)
	}

	key, val, ok = c.GetOldest()
	if !ok || key != "second" || val != 2 {
		t.Errorf("expected oldest to be ('second', 2), got (%v, %v)", key, val)
	}
}

func TestRemoveOldestEmptyCache(t *testing.T) {
	c, _ := New(100, 10)

	key, value, ok := c.RemoveOldest()
	if ok {
		t.Errorf("expected RemoveOldest to return false for empty cache, got true")
	}
	if key != nil {
		t.Errorf("expected nil key for empty cache, got %v", key)
	}
	if value != nil {
		t.Errorf("expected nil value for empty cache, got %v", value)
	}
}

func TestGetOldestEmptyCache(t *testing.T) {
	c, _ := New(100, 10)

	key, value, ok := c.GetOldest()
	if ok {
		t.Errorf("expected GetOldest to return false for empty cache, got true")
	}
	if key != nil {
		t.Errorf("expected nil key for empty cache, got %v", key)
	}
	if value != nil {
		t.Errorf("expected nil value for empty cache, got %v", value)
	}
}

func TestOrderAfterAccess(t *testing.T) {
	c, _ := New(100, 10)
	c.Add("a", "A", 1)
	c.Add("b", "B", 1)
	c.Add("c", "C", 1)
	key, _, _ := c.GetOldest()
	if key != "a" {
		t.Errorf("expected oldest key 'a', got %v", key)
	}
	_, _ = c.Get("a")
	key, _, _ = c.GetOldest()
	if key != "b" {
		t.Errorf("expected oldest key 'b' after access, got %v", key)
	}
}

func TestKeysOrdering(t *testing.T) {
	c, _ := New(100, 10)
	c.Add("a", "A", 1)
	c.Add("b", "B", 1)
	c.Add("c", "C", 1)
	_, _ = c.Get("b")
	keys := c.Keys()
	expected := []interface{}{"a", "c", "b"}
	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(keys))
	}
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("at index %d: expected key %v, got %v", i, expected[i], key)
		}
	}
}

func TestTotalAndWeight(t *testing.T) {
	c, _ := New(100, 10)
	c.Add("a", "A", 5)
	c.Add("b", "B", 10)
	w, n := c.Total()
	if w != 15 {
		t.Errorf("expected total weight 15, got %d", w)
	}
	if n != 2 {
		t.Errorf("expected total items 2, got %d", n)
	}
}

func TestResize(t *testing.T) {
	c, _ := New(50, 5)
	c.Add("a", 1, 10)
	c.Add("b", 2, 10)
	c.Add("c", 3, 10)
	if c.Len() != 3 {
		t.Errorf("expected 3 items, got %d", c.Len())
	}
	evicted := c.Resize(15, 2) // maxWeight now 15, maxSize 2.
	if evicted == 0 {
		t.Errorf("expected evictions due to resize, got %d", evicted)
	}
	if c.Weight() > 15 {
		t.Errorf("expected total weight <= 15, got %d", c.Weight())
	}
	if c.Len() > 2 {
		t.Errorf("expected cache length <= 2, got %d", c.Len())
	}
}
