package ringbuffer

const minQueueLen = 2

// 先进先出
type Queue struct {
	buf               []interface{}
	bufLen            int
	head, tail, count int
}

// 容量是2的n次幂,非线程安全,适当的初始容量大小可以减少内存分配次数
// 对比QueArray有一倍左右的性能提升
func NewQueue() *Queue {
	return &Queue{buf: make([]interface{}, minQueueLen), bufLen: minQueueLen}
}

func (q *Queue) Length() int {
	return q.count
}

// 2倍缩放
func (q *Queue) resize() {
	if q.count < minQueueLen {
		return
	}

	if q.count == q.bufLen || (q.count<<2) == q.bufLen {
		newBuf := make([]interface{}, q.count<<1)
		if q.tail > q.head {
			copy(newBuf, q.buf[q.head:q.tail])
		} else {
			n := copy(newBuf, q.buf[q.head:])
			copy(newBuf[n:], q.buf[:q.tail])
		}
		q.head = 0
		q.tail = q.count
		q.buf = newBuf
		q.bufLen = len(newBuf)
	}
}

// 追加到队尾
func (q *Queue) PushBack(elem interface{}) {
	q.resize()

	q.buf[q.tail] = elem
	q.tail = (q.tail + 1) & (q.bufLen - 1)
	q.count++
}

// 弹出队头
func (q *Queue) PopFront() interface{} {
	if q.count <= 0 {
		return nil
	}
	ret := q.buf[q.head]
	q.buf[q.head] = nil
	q.head = (q.head + 1) & (q.bufLen - 1)
	q.count--

	q.resize()
	return ret
}

//返回队头
func (q *Queue) Front() interface{} {
	if q.count <= 0 {
		return nil
	}
	return q.buf[q.head]
}

//返回队头
func (q *Queue) Back() interface{} {
	if q.count <= 0 {
		return nil
	}
	return q.Get(q.count - 1)
}

// 正数队头遍历，负数队尾遍历
func (q *Queue) Get(i int) interface{} {
	if i < 0 {
		i += q.count
	}

	if i < 0 || i >= q.count {
		return nil
	}
	return q.buf[(q.head+i)&(q.bufLen-1)]
}

type QueArray struct {
	array []interface{}
}

func (s *QueArray) Push(elem interface{}) {
	s.array = append(s.array, elem)
}

func (s *QueArray) Pop() interface{} {
	if len(s.array) > 0 {
		ret := s.array[0]
		s.array = s.array[1:]
		return ret
	}
	return nil
}
