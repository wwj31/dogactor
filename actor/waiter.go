package actor

import (
	"github.com/wwj31/dogactor/timer"
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
	canStop bool
	Base
}

func (s *waiter) OnInit() {}

func (s *waiter) OnHandle(msg Message) {
	switch data := msg.Payload().(type) {
	case *requestWait:
		s.Request(data.targetId, data.msg, data.timeout).Handle(func(resp any, er error) {
			go func() {
				deadline := timer.Get(10 * time.Second)
				select {
				case data.response <- result{data: resp, err: er}:
				case <-deadline.C:
					log.SysLog.Warnw("waiter result put time out sourceId:%v targetId:%v", msg.GetSourceId(), data.targetId)
					break
				}
				timer.Put(deadline)
			}()
		})

	case string:
		if data == "stop" {
			s.canStop = true
			s.Exit()
		} else {
			log.SysLog.Errorw("no such case type", "t", reflect.TypeOf(msg).Name(), "str", data)
		}
	}
}

func (s *waiter) OnStop() bool {
	return s.canStop
}
