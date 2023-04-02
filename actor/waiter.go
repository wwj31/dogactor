package actor

import (
	"reflect"
	"time"

	"github.com/wwj31/dogactor/log"
)

// WARN: Deadlock cause by multiple actor RequestWait to each other

type (
	requestWait struct {
		targetId Id
		timeout  time.Duration
		msg      interface{}
		response chan result
	}
)

type waiter struct {
	Base
}

func (s *waiter) OnInit() {}

func (s *waiter) OnHandle(msg Message) {
	switch data := msg.RawMsg().(type) {
	case *requestWait:
		s.Request(data.targetId, data.msg, data.timeout).Handle(func(resp any, er error) {
			go func() {
				deadline := globalTimerPool.Get(10 * time.Second)
				select {
				case data.response <- result{data: resp, err: er}:
				case <-deadline.C:
					log.SysLog.Warnw("waiter result put time out sourceId:%v targetId:%v", msg.GetSourceId(), data.targetId)
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
