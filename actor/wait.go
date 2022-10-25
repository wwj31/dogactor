package actor

import (
	"github.com/wwj31/dogactor/log"
	"reflect"
	"time"
)

// WARN: Deadlock cause by multiple actor RequestWait to each other

type (
	result struct {
		result interface{}
		err    error
	}

	requestWait struct {
		targetId string
		timeout  time.Duration
		msg      interface{}
		c        chan result
	}
)

type waitActor struct {
	Base
}

func (s *waitActor) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	switch data := msg.(type) {
	case *requestWait:
		req := s.Request(data.targetId, data.msg, data.timeout)
		req.Handle(func(resp interface{}, e error) {
			go func() {
				select {
				case data.c <- result{result: resp, err: e}:
				case <-time.After(10 * time.Second):
					log.SysLog.Warnw("waitActor result put time out sourceId:%v targetId:%v", sourceId, data.targetId)
					break
				}
			}()
		})
	default:
		if str, ok := msg.(string); ok && str == "stop" {
			s.Exit()
		} else {
			log.SysLog.Errorw("no such case type", "t", reflect.TypeOf(msg).Name(), "str", str)
		}
	}
}

func (s *waitActor) OnStop() bool {
	return false
}
