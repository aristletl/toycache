package toycache

import (
	"context"
	lru "github.com/hashicorp/golang-lru"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type MaxMemoryCache struct {
	Cache
	sync.Locker
	max  int64
	used int64
	lru  *lru.Cache
}

func NewMaxMemoryCache(max int64, cache Cache) *MaxMemoryCache {
	res := &MaxMemoryCache{
		max:   max,
		Cache: cache,
	}

	l, _ := lru.New(int(max))
	res.lru = l
	res.Cache.OnEvicted(func(key string, val []byte) {
		// 注册回调
		ok := l.Remove(key)
		if ok {
			atomic.AddInt64(&res.used, -int64(len(val)))
		}
	})
	return res
}

func (m *MaxMemoryCache) Set(ctx context.Context, key string, val []byte,
	expiration time.Duration) error {
	// 在这里判断内存使用量，以及腾出空间
	valSize := int64(len(val))
	for atomic.LoadInt64(&m.used)+valSize > m.max {
		k, _, ok := m.lru.RemoveOldest()
		if ok {
			err := m.Delete(ctx, k.(string))
			if err != nil {
				return err
			}
		} else {
			log.Println("无法移除")
			break
		}
	}
	m.lru.Add(key, valSize)
	atomic.AddInt64(&m.used, valSize)
	return m.Cache.Set(ctx, key, val, expiration)
}
