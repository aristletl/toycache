package toycache

import "time"

// RetryStrategy 重试策略
type RetryStrategy interface {
	// Next 返回重试的时间间隔以及是否需要重试
	Next() (time.Duration, bool)
}

type fixIntervalRetry struct {
	expiration time.Duration
	max        int
	cnt        int
}

func (f *fixIntervalRetry) Next() (time.Duration, bool) {
	f.cnt++
	return f.expiration, f.cnt < f.max
}
