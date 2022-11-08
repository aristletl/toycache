package toycache

import "testing"

func TestLRU_Add(t *testing.T) {
	testCase := []struct {
		name  string
		key   string
		value string

		wantList *node
	}{
		{
			name: "value 为空",
			key:  "nil",
			wantList: &node{
				key:   "nil",
				value: AnyValue{},
				pre:   nil,
				next: &node{
					key:   nil,
					value: AnyValue{},
					pre:   nil,
					next:  nil,
				},
			},
		},
	}
}
