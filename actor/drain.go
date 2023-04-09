package actor

import (
	"github.com/wwj31/dogactor/actor/internal"
	"github.com/wwj31/dogactor/log"
	"reflect"
	"sync/atomic"
	"time"
)

// the drain mode is a safe exit mode provided for
// the purpose of securely and seamlessly transitioning an actor from
// one system to another for continued run.

type drain struct {
	*Base

	draining     atomic.Value
	afterDrained []func()
}

type drained struct{}

func newDrain(base *Base) *drain {
	dr := &drain{Base: base}

	dr.draining.Store(false)

	dr.AppendHandler(func(message Message) bool {
		if _, ok := message.RawMsg().(drained); ok {
			dr.drained()
			return false
		}
		return true
	})
	return dr
}

func (s *drain) Drain(afterDrained ...func()) {
	s.Request(s.System().clusterId, &internal.ReqMsgDrain{}, 5*time.Minute).Handle(func(resp any, err error) {
		if err != nil {
			log.SysLog.Errorw("drain failed ", "err", err)
			return
		}

		respMsgDrain, ok := resp.(*internal.RespMsgDrain)
		if !ok {
			log.SysLog.Errorw("drain resp failed  ", "resp", reflect.TypeOf(resp).String())
			return
		}

		if respMsgDrain.Err != nil {
			log.SysLog.Errorw("drain return err ", "err", respMsgDrain.Err)
			return
		}

		s.afterDrained = afterDrained
		_ = s.Send(s.ID(), drained{})
		s.draining.Store(true)
	})
}

func (s *drain) Draining() bool {
	return s.draining.Load() == true
}

func (s *drain) drained() {
	for _, fn := range s.afterDrained {
		fn()
	}
	s.Exit()
}
