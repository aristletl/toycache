package toycache

import (
	"context"
	"log"
	"sync/atomic"
	"time"
)

type MaxMemoryCache struct {
	Cache
	max         int64
	safeLine    int64
	used        int64
	elimination EliminateStrategy
}

func NewMaxMemoryCache(max, safeLine int64, e EliminateStrategy, cache Cache) *MaxMemoryCache {
	res := &MaxMemoryCache{
		max:         max,
		safeLine:    safeLine,
		Cache:       cache,
		elimination: e,
	}
	res.Cache.OnEvicted(func(key string, val []byte) {
		// 注册回调
		ok := res.elimination.Remove(key)
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
	if atomic.LoadInt64(&m.used)+valSize > m.max {
		for atomic.LoadInt64(&m.used)+valSize > m.safeLine {
			k, _, ok := m.elimination.GetEliminatedKey()
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
	}
	m.elimination.Add(key, valSize)
	atomic.AddInt64(&m.used, valSize)
	return m.Cache.Set(ctx, key, val, expiration)
}
