package queue

import (
	"container/list"
	"fmt"
	"sync"
)

type QueueData interface {
	Key(interface{}) interface{}
	String(interface{}) string
}
type node struct {
	d QueueData
}

type QueueX struct {
	sync.Locker
	datas map[interface{}]*node
	list  *list.List
}

func NewQueueX() *QueueX {
	return &QueueX{
		Locker: newSpinLock(),
		datas:  make(map[interface{}]*node),
		list:   list.New(),
	}
}

func (q *QueueX) Length() int {
	return len(q.datas)
}

func (q *QueueX) PushBack(d ...QueueData) error {
	if len(d) == 0 {
		return nil
	}
	q.Lock()
	defer q.Unlock()

	return q.pushBack(d...)
}

func (q *QueueX) ReplaceOrPushBack(d ...QueueData) {
	if len(d) == 0 {
		return
	}
	q.Lock()
	defer q.Unlock()

	for _, data := range d {
		if e, ok := q.datas[data.Key(data)]; ok {
			e.d = data
		} else {
			q.pushBack(data)
		}
	}

}

// 弹出c 个
func (q *QueueX) PopFront(c int) []QueueData {
	if c == 0 {
		return nil
	}
	q.Lock()
	defer q.Unlock()

	if c < 0 {
		c = q.Length()
	}
	ret := []QueueData{}
	for i := 0; i < c; i++ {
		f := q.list.Front()
		if f == nil {
			return ret
		}
		n := f.Value.(*node)
		q.list.Remove(f)
		delete(q.datas, n.d.Key(n.d))
		ret = append(ret, n.d)
	}
	return ret
}

func (q *QueueX) pushBack(d ...QueueData) error {
	for _, data := range d {
		n := &node{d: data}
		key := data.Key(data)
		if q.datas[key] != nil {
			return fmt.Errorf("exist k:%v v:%v", key, data.String(data))
		}
		q.datas[key] = n
		q.list.PushBack(n)
	}
	return nil
}
