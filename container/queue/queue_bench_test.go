package queue

import (
	"github.com/google/uuid"
	"testing"
)

type T struct {
	s string
}

func (s *T) Key(interface{}) interface{} {
	return s.s
}
func (s *T) String(interface{}) string {
	return s.s
}
func BenchmarkNewQueueX(b *testing.B) {
	queue := NewQueueX()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			queue.PushBack(&T{s: uuid.New().String()}, &T{s: uuid.New().String()})
			//queue.ReplaceOrPushBack(s, s)
			//queue.PopFront()
		}
	})
}
