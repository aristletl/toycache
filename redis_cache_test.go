package toycache

import (
	"context"
	"github.com/aristletl/toycache/mocks"
	"github.com/go-redis/redis/v9"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRedisCache_Set(t *testing.T) {
	ctrl := gomock.NewController(t)
	testCase := []struct {
		name       string
		cmd        func() redis.Cmdable
		key        string
		val        any
		expiration time.Duration
		wantErr    error
	}{
		{
			name: "return ok",
			cmd: func() redis.Cmdable {
				res := mocks.NewMockCmdable(ctrl)
				cmd := redis.NewStatusCmd(nil)
				cmd.SetVal("OK")
				res.EXPECT().Set(gomock.Any(), "key", "value", time.Minute).
					Return(cmd)
				return res
			},
			key:        "key",
			val:        "value",
			expiration: time.Minute,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			client := NewRedisCache(tc.cmd())
			err := client.Set(context.Background(), tc.key, tc.val, tc.expiration)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
