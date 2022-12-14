package toycache

import (
	"context"
	"errors"
	"time"
)

// 值的问题
// - string：可以，问题是本地缓存，结构体转化为 string，比如 json 表达 User
// - []byte：最通用的表达， 可以存储序列化后的数据，也可以存储加密数据，还可以存储压缩数据。用户用起来不方便
// - any：Redis 之类的实现，但是要考虑序列化的问题
type Cache interface {
	// Get(ctx context.Context, key string) (AnyValue, error)
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, val any, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	OnEvicted(func(key string, val []byte))
}

type EliminateStrategy interface {
	// GetOldest 获取应该被淘汰的key
	GetEliminatedKey() (any, AnyValue, bool)
	// Add 添加一个key
	Add(key any, args ...any)

	Remove(key any) bool

	Get(key any) (AnyValue, bool)
}

type AnyValue struct {
	Val any
	Err error
}

func (a AnyValue) String() (string, error) {
	if a.Err != nil {
		return "", a.Err
	}
	str, ok := a.Val.(string)
	if !ok {
		return "", errors.New("无法转换的类型")
	}
	return str, nil
}

func (a AnyValue) Int64() (int64, error) {
	if a.Err != nil {
		return 0, a.Err
	}
	res, ok := a.Val.(int64)
	if !ok {
		return -1, errors.New("无法转换的类型")
	}
	return res, nil
}
