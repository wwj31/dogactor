package actor

import (
	"reflect"
	"time"

	"github.com/wwj31/dogactor/log"
)

// WARN: Deadlock cause by multiple actor RequestWait to each other

type (
	requestWait struct {
		targetId string
		timeout  time.Duration
		msg      interface{}
		c        chan result
	}
)

type waiter struct {
	Base
}

func (s *waiter) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	switch data := msg.(type) {
	case *requestWait:
		req := s.Request(data.targetId, data.msg, data.timeout)
		req.Handle(func(resp interface{}, e error) {
			go func() {
				deadline := globalTimerPool.Get(10 * time.Second)
				select {
				case data.c <- result{data: resp, err: e}:
				case <-deadline.C:
					log.SysLog.Warnw("waiter result put time out sourceId:%v targetId:%v", sourceId, data.targetId)
					break
				}
				globalTimerPool.Put(deadline)
			}()
		})

	case string:
		if data == "stop" {
			s.Exit()
		} else {
			log.SysLog.Errorw("no such case type", "t", reflect.TypeOf(msg).Name(), "str", data)
		}
	}
}

func (s *waiter) OnStop() bool {
	return false
}
