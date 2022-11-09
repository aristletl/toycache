package toycache

import "sync"

type LRU struct {
	sync.RWMutex
	head  *node
	tail  *node
	cache map[any]*node
}

func NewLRU() *LRU {
	head, tail := new(node), new(node)
	head.next = tail
	tail.pre = head
	return &LRU{
		head:  head,
		tail:  tail,
		cache: make(map[any]*node),
	}
}

func (l *LRU) Get(key any) (AnyValue, bool) {
	l.RLock()
	val, ok := l.cache[key]
	l.RUnlock()
	if !ok {
		return AnyValue{}, false
	}
	return AnyValue{Val: val.value}, true
}

func (l *LRU) GetEliminatedKey() (any, AnyValue, bool) {
	if len(l.cache) == 0 {
		return nil, AnyValue{}, false
	}

	l.RLock()
	res := l.tail.pre
	l.RUnlock()
	return res.key, AnyValue{Val: res.value}, true
}

func (l *LRU) Add(key any, args ...any) {
	if key == nil {
		return
	}
	var val any
	if len(args) != 0 {
		val = args[0]
	}

	l.Lock()
	defer l.Unlock()
	n, ok := l.cache[key]
	if ok {
		n.value = val
		remove(n)
	} else {
		n = &node{
			key:   key,
			value: val,
		}
		l.cache[key] = n
	}
	insert(l.head, n)
}

func (l *LRU) Remove(key any) bool {
	l.RLock()
	n, ok := l.cache[key]
	l.RUnlock()
	if !ok {
		return true
	}

	l.Lock()
	delete(l.cache, key)
	remove(n)
	l.Unlock()
	return true
}

type node struct {
	key   any
	value any
	pre   *node
	next  *node
}

// insert the new node behind the pre
func insert(pre, newNode *node) {
	if pre.next != nil {
		pre.next.pre = newNode
	}
	newNode.next = pre.next
	pre.next = newNode
	newNode.pre = pre
}

func remove(p *node) {
	p.pre.next = p.next
	p.next.pre = p.pre
}
