package queue

import (
	"runtime"
	"sync/atomic"
)

type spinLock struct {
	lock int32
}

func newSpinLock() *spinLock {
	return &spinLock{}
}

func (s *spinLock) Lock() {
	for !atomic.CompareAndSwapInt32(&s.lock, 0, 1) {
		runtime.Gosched()
	}
}
func (s *spinLock) Unlock() {
	atomic.CompareAndSwapInt32(&s.lock, 1, 0)
}
