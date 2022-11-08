package toycache

import (
	"context"
	_ "embed"
	"errors"
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
	client     redis.Cmdable
	value      string
	key        string
	expiration time.Duration

	unlock     chan struct{}
	unlockOnce sync.Once
}

func NewClient(cmd redis.Cmdable) *Client {
	return &Client{
		client: cmd,
		unlock: make(chan struct{}),
	}
}

func (c *Client) Lock(ctx context.Context, key string, expiration time.Duration) error {
	c.value = uuid.New().String()

	for {
		res, err := c.client.Eval(ctx, lock, []string{key}, c.value, expiration.Seconds()).Result()
		if res == "OK" {
			return nil
		}
		if err == context.DeadlineExceeded {

		}
		if err != nil {
			return err
		}
	}
}

func (c *Client) TryLock(ctx context.Context, key string, expiration time.Duration) error {
	c.key = key
	c.value = uuid.New().String()
	c.expiration = expiration
	ok, err := c.client.SetNX(ctx, key, c.value, expiration).Result()
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("redis-lock: 抢锁失败")
	}
	return nil
}

func (c *Client) UnLock(ctx context.Context, key string) error {
	c.unlockOnce.Do(func() {
		close(c.unlock)
	})
	// 解锁，我们需要先判断这个锁是不是我们持有的，即 get 这个 key 的值
	// 是否是自己所设置的值，如果是的话，我们就把它删除。但是，get 与 del
	// 在 redis 中是两个命令，非原子性的，所以，我们需要考虑使用lua脚本来保证原子性
	res, err := c.client.Eval(ctx, unlock, []string{c.key}, c.value).Int64()
	if err != nil {
		return err
	}

	if res != 1 {
		return errs.ErrLockNotHold
	}
	return nil
}

func (c *Client) AutoRefresh(internal time.Duration, timeout time.Duration) error {

	retrySignal := make(chan struct{}, 1)
	defer close(retrySignal)
	ticker := time.NewTicker(internal)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := c.Refresh(ctx)
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
			err := c.Refresh(ctx)
			cancel()
			if err == context.DeadlineExceeded {
				retrySignal <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}
		case <-c.unlock:
			return nil
		}
	}
}

func (c *Client) Refresh(ctx context.Context) error {
	res, err := c.client.Eval(ctx, refresh, []string{c.key}, c.value, c.expiration.Seconds()).Int64()
	//if err == redis.Nil {
	//	return errs.ErrKeyNotFound
	//}
	if err != nil {
		return err
	}

	if res != 1 {
		return errors.New("redis-lock: ")
	}
	return nil
}
