package actor

import (
	"time"
)

//谨慎使用，可能带来死锁问题
type waitActor struct {
	Base

	targetId string
	timeout  time.Duration
	msg      interface{}
	c        chan result
}
type result struct {
	result interface{}
	err    error
}

func (s *waitActor) OnInit() {
	//发出请求，等待结果
	req := s.Request(s.targetId, s.msg, s.timeout)
	req.Handle(func(resp interface{}, e error) {
		s.c <- result{result: resp, err: e}
		s.Exit()
	})
}
func (s *waitActor) OnStop() bool {
	close(s.c)
	return true
}
