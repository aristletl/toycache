package errs

import (
	"errors"
)

var (
	ErrKeyNotFound = errors.New("cache: 找不到key")

	ErrLockNotHold = errors.New("redis-lock: 未占用锁")
)
