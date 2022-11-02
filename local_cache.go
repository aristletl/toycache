package toycache

import (
	"context"
	"errors"
	"github.com/aristletl/toycache/internal/errs"
	"sync"
	"time"
)

type LocalCache struct {
	data map[string]item
	sync.RWMutex
}

func (l *LocalCache) Get(ctx context.Context, key string) (any, error) {
	l.RLock()
	val, ok := l.data[key]
	l.RUnlock()
	if !ok {
		return nil, errs.NewErrKeyNotFound(key)
	}

	if val.deadline.Before(time.Now()) {
		return nil, errors.New("")
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
