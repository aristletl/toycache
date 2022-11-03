package toycache

import (
	"context"
	"github.com/aristletl/toycache/internal/errs"
	"sync"
	"time"
)

type LocalCache struct {
	data map[string]item
	sync.RWMutex
}

func NewLocalCache() *LocalCache {
	c := &LocalCache{
		data: make(map[string]item),
	}

	go func() {
		timer := time.NewTicker(10*time.Second)
		for {
			select {
			case <- timer.C:
			case <-:
				
			}
		}
	}()

	return c
}

func (l *LocalCache) Get(ctx context.Context, key string) (any, error) {
	l.RLock()
	val, ok := l.data[key]
	l.RUnlock()
	if !ok {
		return nil, errs.NewErrKeyNotFound(key)
	}

	// double check
	now := time.Now()
	if val.deadline.Before(now) {
		l.Lock()
		defer l.Unlock()
		val, ok = l.data[key]
		if !ok {
			return nil, errs.NewErrKeyNotFound(key)
		}
		if val.deadline.Before(now) {
			delete(l.data, key)
			return nil, errs.NewErrKeyNotFound(key)
		}
	}
	return val.val, nil
}

func (l *LocalCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	//TODO implement me
	panic("implement me")
}

func (l *LocalCache) Delete(ctx context.Context, key string) error {
	//TODO implement me
	panic("implement me")
}

type item struct {
	val      any
	deadline time.Time
}
