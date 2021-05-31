package lru

import (
	"container/list"
)

type node struct {
	k interface{}
	v interface{}
}

type LRU struct {
	maxLen  int // if maxLen<=0 then no limit
	lruList *list.List
	datas   map[interface{}]*list.Element
}

func NewLRU(maxLen int) *LRU {
	return &LRU{
		maxLen:  maxLen,
		lruList: list.New(),
		datas:   make(map[interface{}]*list.Element),
	}
}

func (lc *LRU) PushFront(k, v interface{}) {
	if e, ok := lc.datas[k]; ok {
		e.Value.(*node).v = v
		lc.lruList.MoveToFront(e)
	} else {
		e := lc.lruList.PushFront(&node{k, v})
		lc.datas[k] = e
	}
	if lc.maxLen > 0 && lc.lruList.Len() > lc.maxLen {
		lc.PopTail()
	}
}

func (lc *LRU) PopTail() (interface{}, bool) {
	e := lc.lruList.Back()
	if e != nil {
		lc.Del(e.Value.(*node).k)
		return e.Value.(*node).v, true
	}
	return nil, false
}

func (lc *LRU) Del(k interface{}) {
	if e, ok := lc.datas[k]; ok {
		delete(lc.datas, k)
		lc.lruList.Remove(e)
	}
}

func (lc *LRU) Get(k interface{}) (interface{}, bool) {
	if e, ok := lc.datas[k]; ok {
		lc.lruList.MoveToFront(e)
		return e.Value.(*node).v, ok
	}
	return nil, false
}

func (lc *LRU) Len() int {
	return lc.lruList.Len()
}
