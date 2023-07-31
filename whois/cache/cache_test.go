package cache

import (
	"testing"
)

func TestCache(t *testing.T) {
    lru := NewLRUCache[int,int](5)

    lru.Put(1, 5)
    lru.Put(2, 9)
    lru.Put(3, 9)
    lru.Put(4, 9)
    lru.Put(5, 9)
    lru.Put(6, 9)
    if val, found := lru.Get(6); !found || val != 9 {
		t.Error("not found")
	}
    lru.Put(7, 9)
    if _, found := lru.Get(1); found {
		t.Error("wasnt removed")
	}
    if _, found := lru.Get(7); !found {
		t.Error("not found")
	}

    if len(lru.idx) != 5 {
		t.Error("incorrect length")
	}
}

type Data struct {
    Value int
}

func TestCacheMoveTail(t *testing.T) {
    lru := NewLRUCache[string,Data](2)

    lru.Put("1", Data{})
    lru.Put("2", Data{1})
    lru.Get("1")
    lru.Put("3", Data{})
	/* should be removed since "1" was moved up */
    if _, found := lru.Get("2") ; found {
		t.Error("wasnt removed")
	}
}
