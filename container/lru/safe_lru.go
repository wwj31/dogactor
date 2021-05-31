package lru

import (
	"container/list"
	"sync"
)

type SafeLRU struct {
	mutex sync.RWMutex

	maxLen  int // if maxLen<=0 then no limit
	lruList *list.List
	datas   map[interface{}]*list.Element
}

func NewSafeLRU(maxLen int) *SafeLRU {
	return &SafeLRU{
		maxLen:  maxLen,
		lruList: list.New(),
		datas:   make(map[interface{}]*list.Element),
	}
}

func (lc *SafeLRU) PushFront(k, v interface{}) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	if e, ok := lc.datas[k]; ok {
		e.Value.(*node).v = v
		lc.lruList.MoveToFront(e)
	} else {
		e := lc.lruList.PushFront(&node{k, v})
		lc.datas[k] = e
	}
	if lc.maxLen > 0 && lc.lruList.Len() > lc.maxLen {
		lc.popTail()
	}
}

func (lc *SafeLRU) PopTail() (interface{}, bool) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	return lc.popTail()
}

func (lc *SafeLRU) Get(k interface{}) (interface{}, bool) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	if e, ok := lc.datas[k]; ok {
		lc.lruList.MoveToFront(e)
		return e.Value.(*node).v, ok
	}
	return nil, false
}

func (lc *SafeLRU) Del(k interface{}) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	lc.del(k)
}

func (lc *SafeLRU) Len() int {
	lc.mutex.RLock()
	defer lc.mutex.RUnlock()
	return lc.lruList.Len()
}

func (lc *SafeLRU) popTail() (interface{}, bool) {
	e := lc.lruList.Back()
	if e != nil {
		lc.del(e.Value.(*node).k)
		return e.Value.(*node).v, true
	}
	return nil, false
}

func (lc *SafeLRU) del(k interface{}) {
	if e, ok := lc.datas[k]; ok {
		delete(lc.datas, k)
		lc.lruList.Remove(e)
	}
}
