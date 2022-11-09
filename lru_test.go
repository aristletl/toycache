package toycache

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLRU_Add(t *testing.T) {
	testCase := []struct {
		name     string
		cache    *LRU
		keys     []string
		value    []any
		wantList *node
	}{
		{
			name:  "顺序插入",
			keys:  []string{"1", "2"},
			cache: NewLRU(),
			wantList: &node{
				key:   "2",
				value: AnyValue{},
				pre:   nil,
				next: &node{
					key:   "1",
					value: AnyValue{},
					pre:   nil,
					next:  nil,
				},
			},
		},
		{
			name:  "二次插入",
			keys:  []string{"1", "2", "1"},
			cache: NewLRU(),
			wantList: &node{
				key:   "1",
				value: AnyValue{},
				pre:   nil,
				next: &node{
					key:   "2",
					value: AnyValue{},
					pre:   nil,
					next:  nil,
				},
			},
		},
		{
			name:  "存值",
			cache: NewLRU(),
			keys:  []string{"1", "2"},
			value: []any{1, 2},
			wantList: &node{
				key:   "2",
				value: AnyValue{Val: 2},
				pre:   nil,
				next: &node{
					key:   "1",
					value: AnyValue{Val: 1},
					pre:   nil,
					next:  nil,
				},
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			for i, key := range tc.keys {
				var val any
				if i < len(tc.value) {
					val = tc.value[i]
				}
				tc.cache.Add(key, val)
			}
			listEqual(t, tc.wantList, tc.cache.head.next)
		})
	}
}

func listEqual(t *testing.T, src, dst *node) bool {
	if src == nil || dst == nil {
		return src == dst
	}

	if assert.Equal(t, src.key, dst.key) || assert.Equal(t, src.value, dst.value) {
		return false
	}

	return listEqual(t, src.next, dst.next) && listEqual(t, src.pre, dst.pre)
}

func TestLRU_Get(t *testing.T) {
	testCase := []struct {
		name    string
		cache   *LRU
		key     any
		wantOk  bool
		wantVal AnyValue
	}{
		{
			name:    "获取不存在的值",
			cache:   NewLRU(),
			key:     1,
			wantOk:  false,
			wantVal: AnyValue{},
		},
		{
			name:    "key 为 nil",
			cache:   NewLRU(),
			key:     nil,
			wantOk:  false,
			wantVal: AnyValue{},
		},
		{
			name: "获取以存入的值",
			cache: func() *LRU {
				l := NewLRU()
				l.Add(1, 1)
				return l
			}(),
			key:     1,
			wantOk:  true,
			wantVal: AnyValue{Val: 1},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			res, ok := tc.cache.Get(tc.key)
			assert.Equal(t, tc.wantOk, ok)
			assert.Equal(t, tc.wantVal, res)
		})
	}
}

func TestLRU_Remove(t *testing.T) {
	testCase := []struct {
		name     string
		cache    *LRU
		key      any
		wantRes  bool
		wantList *node
	}{
		{
			name:    "删除不存在的值",
			cache:   NewLRU(),
			key:     1,
			wantRes: true,
		},
		{
			name:    "key 为 nil",
			cache:   NewLRU(),
			wantRes: true,
		},
		{
			name: "",
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.cache.Remove(tc.key)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
