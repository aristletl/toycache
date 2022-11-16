package toycache

import (
	"context"
	"github.com/aristletl/toycache/internal/errs"
	"sync"
	"time"
)

type LocalCacheOption func(l *LocalCache)

type LocalCache struct {
	sync.RWMutex
	data      map[string]*item
	close     chan struct{}
	closeOnce sync.Once

	onEvicted func(key string, val any)
}

func (l *LocalCache) OnEvicted(fn func(key string, val []byte)) {
	panic("implement me")
}

func NewLocalCache(opts ...LocalCacheOption) *LocalCache {
	l := &LocalCache{
		data:  make(map[string]*item),
		close: make(chan struct{}, 1),
	}

	for _, opt := range opts {
		opt(l)
	}

	timer := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-timer.C:
				cnt := 0
				l.Lock()
				for k, v := range l.data {
					if v.deadline.Before(time.Now()) {
						l.delete(k, v.val)
					}
					cnt++
					if cnt >= 1000 {
						break
					}
				}
				l.Unlock()
			case <-l.close:
				return
			}
		}
	}()

	return l
}

func (l *LocalCache) delete(key string, val any) {
	delete(l.data, key)
	if l.onEvicted != nil {
		l.onEvicted(key, val)
	}
}

func (l *LocalCache) Get(ctx context.Context, key string) (any, error) {
	l.RLock()
	val, ok := l.data[key]
	l.RUnlock()
	if !ok {
		return nil, errs.ErrKeyNotFound
	}

	// double check
	now := time.Now()
	if val.deadline.Before(now) {
		l.Lock()
		defer l.Unlock()
		val, ok = l.data[key]
		if !ok {
			return nil, errs.ErrKeyNotFound
		}
		if val.deadline.Before(now) {
			l.delete(key, val.val)
			return nil, errs.ErrKeyNotFound
		}
	}
	return val.val, nil
}

func (l *LocalCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	l.Lock()
	defer l.Unlock()
	l.data[key] = &item{
		val:      val,
		deadline: time.Now().Add(expiration),
	}
	return nil
}

func (l *LocalCache) Delete(ctx context.Context, key string) error {
	l.Lock()
	defer l.Unlock()
	val, ok := l.data[key]
	if !ok {
		return nil
	}
	l.delete(key, val.val)
	return nil
}

func (l *LocalCache) Close() error {
	l.closeOnce.Do(func() {
		l.close <- struct{}{}
		close(l.close)
	})
	return nil
}

func WithOnEvicted(fn func(key string, val any)) LocalCacheOption {
	return func(l *LocalCache) {
		l.onEvicted = fn
	}
}

type item struct {
	val      any
	deadline time.Time
}
