package toycache

import (
	"context"
	_ "embed"
	"github.com/aristletl/toycache/internal/errs"
	"github.com/go-redis/redis/v9"
	"github.com/google/uuid"
	"sync"
	"time"
)

// Redis 分布式锁
// 在 Redis 中， 实现一个分布式锁的起点，就是利用 setnx 命令， 确保
// 可以排他的设置一个键值对

var (
	//go:embed internal/lua/unlock.lua
	unlock string

	//go:embed internal/lua/lock.lua
	lock string

	//go:embed internal/lua/refresh.lua
	refresh string
)

type Client struct {
	client redis.Cmdable
}

func NewClient(cmd redis.Cmdable) *Client {
	return &Client{
		client: cmd,
	}
}

// Lock 重试策略加锁
func (c *Client) Lock(ctx context.Context, key string, expiration time.Duration, retry RetryStrategy) (*Lock, error) {
	val := uuid.New().String()

	for {
		res, err := c.client.Eval(ctx, lock, []string{key}, val, expiration.Seconds()).Result()
		if res == "OK" {
			return newLock(c.client, key, val, expiration), nil
		}
		if err == context.DeadlineExceeded && retry != nil {
			interval, ok := retry.Next()
			if !ok {
				return nil, errs.ErrFailedToPreemptLock
			}
			time.Sleep(interval)
			continue
		}
		if err != nil {
			return nil, err
		}
	}
}

func (c *Client) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	val := uuid.New().String()
	ok, err := c.client.SetNX(ctx, key, val, expiration).Result()
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errs.ErrFailedToPreemptLock
	}
	return newLock(c.client, key, val, expiration), nil
}

type Lock struct {
	client     redis.Cmdable
	value      string
	key        string
	expiration time.Duration

	unlock     chan struct{}
	unlockOnce sync.Once
}

func newLock(client redis.Cmdable, key, val string, expiration time.Duration) *Lock {
	return &Lock{
		client:     client,
		key:        key,
		value:      val,
		expiration: expiration,
		unlock:     make(chan struct{}),
	}
}

func (l *Lock) UnLock(ctx context.Context) error {
	l.unlockOnce.Do(func() {
		close(l.unlock)
	})
	// 解锁，我们需要先判断这个锁是不是我们持有的，即 get 这个 key 的值
	// 是否是自己所设置的值，如果是的话，我们就把它删除。但是，get 与 del
	// 在 redis 中是两个命令，非原子性的，所以，我们需要考虑使用lua脚本来保证原子性
	res, err := l.client.Eval(ctx, unlock, []string{l.key}, l.value).Int64()
	if err != nil {
		return err
	}

	if res != 1 {
		return errs.ErrLockNotHold
	}
	return nil
}

func (l *Lock) AutoRefresh(internal time.Duration, timeout time.Duration) error {

	retrySignal := make(chan struct{}, 1)
	defer close(retrySignal)
	ticker := time.NewTicker(internal)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if err == context.DeadlineExceeded {
				// 进行重试
				retrySignal <- struct{}{}
				continue
			}
			if err != nil {
				// 不可挽回的错误，直接返回
				return err
			}
		case <-retrySignal:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if err == context.DeadlineExceeded {
				retrySignal <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}
		case <-l.unlock:
			return nil
		}
	}
}

func (l *Lock) Refresh(ctx context.Context) error {
	res, err := l.client.Eval(ctx, refresh, []string{l.key}, l.value, l.expiration.Seconds()).Int64()
	//if err == redis.Nil {
	//	return errs.ErrKeyNotFound
	//}
	if err != nil {
		return err
	}

	if res != 1 {
		return errs.ErrLockNotHold
	}
	return nil
}
