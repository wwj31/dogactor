package actor

import (
	"time"
)

// WARN: Deadlock cause by multiple actor RequestWait to each other

type RequestWait struct {
	targetId string
	timeout  time.Duration
	msg      interface{}
	c        chan result
}
type waitActor struct {
	Base
}
type result struct {
	result interface{}
	err    error
}

func (s *waitActor) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	switch data := msg.(type) {
	case *RequestWait:
		req := s.Request(data.targetId, data.msg, data.timeout)
		req.Handle(func(resp interface{}, e error) {
			data.c <- result{result: resp, err: e}
			s.Exit()
		})
	}
}

func (s *waitActor) OnStop() bool {
	return true
}
