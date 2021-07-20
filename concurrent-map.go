package util

import (
	"encoding/json"
	"sync"
)
//参考 https://github.com/orcaman/concurrent-map/blob/master/concurrent_map.go
const SLOTS = 32

type CMap []*CMapSlot

type CMapSlot struct {
	items map[string]interface{}
	sync.RWMutex
}

func New() CMap {
	m := make(CMap, SLOTS)
	for i := 0; i < SLOTS; i++ {
		m[i] = &CMapSlot{items: make(map[string]interface{})}
	}
	return m
}
func (m CMap) GetSlots(key string) *CMapSlot {
	return m[uint(fnv32(key))%uint(SLOTS)]
}
func (m CMap) MSet(data map[string]interface{}) {
	for key, value := range data {
		m.Set(key, value)
	}
}
func (m CMap) Set(key string, value interface{}) {
	slot := m.GetSlots(key)
	slot.Lock()
	slot.items[key] = value
	slot.Unlock()
}

type UpsertCb func(exist bool, valueInMap interface{}, newValue interface{}) interface{}

// Insert or Update - updates existing element or inserts a new one using UpsertCb
func (m CMap) Upsert(key string, value interface{}, cb UpsertCb) (res interface{}) {
	slot := m.GetSlots(key)
	slot.Lock()
	v, ok := slot.items[key]
	res = cb(ok, v, value)
	slot.items[key] = res
	slot.Unlock()
	return res
}

// Sets the given value under the specified key if no value was associated with it.
func (m CMap) SetIfAbsent(key string, value interface{}) bool {
	// Get map shard.
	slot := m.GetSlots(key)
	slot.Lock()
	_, ok := slot.items[key]
	if !ok {
		slot.items[key] = value
	}
	slot.Unlock()
	return !ok
}
func (m CMap) Get(key string) (interface{}, bool) {
	// Get shard
	slot := m.GetSlots(key)
	slot.RLock()
	// Get item from shard.
	val, ok := slot.items[key]
	slot.RUnlock()
	return val, ok
}

func (m CMap) Count() int {
	count := 0
	for i := 0; i < SLOTS; i++ {
		shard := m[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}
func (m CMap) Has(key string) bool {
	// Get shard
	slot := m.GetSlots(key)
	slot.RLock()
	// See if element is within shard.
	_, ok := slot.items[key]
	slot.RUnlock()
	return ok
}
func (m CMap) Remove(key string) {
	// Try to get shard.
	slot := m.GetSlots(key)
	slot.Lock()
	delete(slot.items, key)
	slot.Unlock()
}

type RemoveCb func(key string, v interface{}, exists bool) bool

func (m CMap) RemoveCb(key string, cb RemoveCb) bool {
	// Try to get shard.
	shard := m.GetSlots(key)
	shard.Lock()
	v, ok := shard.items[key]
	remove := cb(key, v, ok)
	if remove && ok {
		delete(shard.items, key)
	}
	shard.Unlock()
	return remove
}
func (m CMap) Pop(key string) (v interface{}, exists bool) {
	// Try to get shard.
	shard := m.GetSlots(key)
	shard.Lock()
	v, exists = shard.items[key]
	delete(shard.items, key)
	shard.Unlock()
	return v, exists
}

// IsEmpty checks if map is empty.
func (m CMap) IsEmpty() bool {
	return m.Count() == 0
}

// Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type Tuple struct {
	Key string
	Val interface{}
}

// Iter returns an iterator which could be used in a for range loop.
//
// Deprecated: using IterBuffered() will get a better performence
func (m CMap) Iter() <-chan Tuple {
	chans := snapshot(m)
	ch := make(chan Tuple)
	go fanIn(chans, ch)
	return ch
}

// IterBuffered returns a buffered iterator which could be used in a for range loop.
func (m CMap) IterBuffered() <-chan Tuple {
	chans := snapshot(m)
	total := 0
	for _, c := range chans {
		total += cap(c)
	}
	ch := make(chan Tuple, total)
	go fanIn(chans, ch)
	return ch
}

// Clear removes all items from map.
func (m CMap) Clear() {
	for item := range m.IterBuffered() {
		m.Remove(item.Key)
	}
}

// Returns a array of channels that contains elements in each shard,
// which likely takes a snapshot of `m`.
// It returns once the size of each buffered channel is determined,
// before all the channels are populated using goroutines.
func snapshot(m CMap) (chans []chan Tuple) {
	chans = make([]chan Tuple, SLOTS)
	wg := sync.WaitGroup{}
	wg.Add(SLOTS)
	// Foreach shard.
	for index, shard := range m {
		go func(index int, shard *CMapSlot) {
			// Foreach key, value pair.
			shard.RLock()
			chans[index] = make(chan Tuple, len(shard.items))
			wg.Done()
			for key, val := range shard.items {
				chans[index] <- Tuple{key, val}
			}
			shard.RUnlock()
			close(chans[index])
		}(index, shard)
	}
	wg.Wait()
	return chans
}

// fanIn reads elements from channels `chans` into channel `out`
func fanIn(chans []chan Tuple, out chan Tuple) {
	wg := sync.WaitGroup{}
	wg.Add(len(chans))
	for _, ch := range chans {
		go func(ch chan Tuple) {
			for t := range ch {
				out <- t
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()
	close(out)
}

// Items returns all items as map[string]interface{}
func (m CMap) Items() map[string]interface{} {
	tmp := make(map[string]interface{})

	// Insert items to temporary map.
	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}

	return tmp
}

// Iterator callback,called for every key,value found in
// maps. RLock is held for all calls for a given shard
// therefore callback sess consistent view of a shard,
// but not across the shards
type IterCb func(key string, v interface{})

// Callback based iterator, cheapest way to read
// all elements in a map.
func (m CMap) IterCb(fn IterCb) {
	for idx := range m {
		shard := (m)[idx]
		shard.RLock()
		for key, value := range shard.items {
			fn(key, value)
		}
		shard.RUnlock()
	}
}

// Keys returns all keys as []string
func (m CMap) Keys() []string {
	count := m.Count()
	ch := make(chan string, count)
	go func() {
		// Foreach shard.
		wg := sync.WaitGroup{}
		wg.Add(SLOTS)
		for _, shard := range m {
			go func(shard *CMapSlot) {
				// Foreach key, value pair.
				shard.RLock()
				for key := range shard.items {
					ch <- key
				}
				shard.RUnlock()
				wg.Done()
			}(shard)
		}
		wg.Wait()
		close(ch)
	}()

	// Generate keys
	keys := make([]string, 0, count)
	for k := range ch {
		keys = append(keys, k)
	}
	return keys
}

//Reviles ConcurrentMap "private" variables to json marshal.
func (m CMap) MarshalJSON() ([]byte, error) {
	// Create a temporary map, which will hold all item spread across shards.
	tmp := make(map[string]interface{})

	// Insert items to temporary map.
	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}
	return json.Marshal(tmp)
}
func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(key)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}
