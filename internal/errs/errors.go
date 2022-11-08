package errs

import (
	"errors"
)

var (
	ErrKeyNotFound = errors.New("cache: 找不到key")
)

var (
	ErrLockNotHold         = errors.New("redis-lock: 未占用锁")
	ErrFailedToPreemptLock = errors.New("redis-lock: 抢锁失败")
)
