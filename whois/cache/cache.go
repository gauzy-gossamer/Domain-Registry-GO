package cache

import (
   "container/list"
   "sync"
)

type CacheVal[K comparable, V any] struct {
    value V
    key K
}

type LRUCache[K comparable, V any] struct {
    capacity int
    list *list.List
    idx map[K]*list.Element
    mu sync.RWMutex
}

func NewLRUCache[K comparable, V any](capacity int) LRUCache[K,V] {
    values := list.New()
    idx := make(map[K]*list.Element, capacity)
    return LRUCache[K,V] {
            list:values,
            idx:idx,
            capacity:capacity,
    }
}

func (lru *LRUCache[K,V]) Get(key K) (V, bool) {
    lru.mu.RLock()
    defer lru.mu.RUnlock()

    node, ok := lru.idx[key]
    if !ok {
        var none V
        return none, false
    }
    lru.list.MoveToFront(node)
    return node.Value.(*CacheVal[K,V]).value, true
}

func (lru *LRUCache[K,V]) Put(key K, value V)  {
    lru.mu.Lock()
    defer lru.mu.Unlock()

    node, ok := lru.idx[key]
    if ok {
        node.Value.(*CacheVal[K,V]).value = value
        lru.list.MoveToFront(node)
        return
    }

    if lru.list.Len() == lru.capacity {
        node = lru.list.Back()
        lru.list.Remove(node)
        delete(lru.idx, node.Value.(*CacheVal[K,V]).key)
    }
    node = lru.list.PushFront(&CacheVal[K,V]{value:value, key:key})
    lru.idx[key] = node
}

func (lru *LRUCache[K,V]) Remove(key K) bool {
    lru.mu.Lock()
    defer lru.mu.Unlock()

    node, ok := lru.idx[key]
    if !ok {
        return false
    }
    lru.list.Remove(node)
    delete(lru.idx, node.Value.(*CacheVal[K,V]).key)
    return true
}