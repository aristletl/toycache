package errs

import (
	"errors"
)

var (
	ErrKeyNotFound    = errors.New("cache: 找不到key")
	ErrFailedToSetKey = errors.New("cache: 设置失败")
	ErrFailedToDelKey = errors.New("cache: 删除失败")
)

var (
	ErrLockNotHold         = errors.New("redis-lock: 未占用锁")
	ErrFailedToPreemptLock = errors.New("redis-lock: 抢锁失败")
)
