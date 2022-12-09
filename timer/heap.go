package timer

type timerHeap []*timer

func (t *timerHeap) Len() int {
	return len(*t)
}

func (t *timerHeap) Less(i, j int) bool {
	arr := *t
	return arr[i].endAt.Before(arr[j].endAt)
}

func (t *timerHeap) Swap(i, j int) {
	arr := *t
	arr[i].index, arr[j].index = j, i
	arr[i], arr[j] = arr[j], arr[i]
}

func (t *timerHeap) Push(v interface{}) {
	_timer := v.(*timer)
	_timer.index = t.Len()
	*t = append(*t, _timer)

}

func (t *timerHeap) Pop() (v interface{}) {
	if t.Len() == 0 {
		return nil
	}
	v = (*t)[0]
	*t = (*t)[:t.Len()-1]
	return v
}

func (t *timerHeap) peek() *timer {
	if t.Len() == 0 {
		return nil
	}
	return (*t)[0]
}
