package toycache

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v9"
	"time"
)

type RedisCache struct {
	client redis.Cmdable
}

//go:generate mockgen -package mocks -destination=mocks/redis_cmdable.mock.go github.com/go-reids/redis/v9 Cmdable
func NewRedisCache(client redis.Cmdable) *RedisCache {
	return &RedisCache{
		client: client,
	}
}

func (r *RedisCache) Get(ctx context.Context, key string) (any, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	res, err := r.client.Set(ctx, key, val, expiration).Result()
	if err != nil {
		return err
	}
	if res != "OK" {
		return errors.New("cache: 设置失败")
	}
	return nil
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	res, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if res != 0 {
		return errors.New("cache: 删除失败")
	}
	return nil
}
