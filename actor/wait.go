package actor

import (
	"time"
)

// WARN: Deadlock cause by multiple actor RequestWait to each other

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
