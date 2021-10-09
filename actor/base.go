package actor

import (
	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/log"
)

// Base 创建Actor统一通过Base继承
type Base struct {
	Actor
}

func (s *Base) initActor(actor Actor) {
	s.Actor = actor
}

// 默认方法，需要的子类重写
func (s *Base) OnInit()      { log.KV("actor", s.GetID()).Warn("actor default init") }
func (s *Base) OnStop() bool { return true }

func (s *Base) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	logger.KV("actorId", s.GetID()).Warn("not implement OnHandleMessage")
}

func (s *Base) OnHandleRequest(sourceId, targetId, requestId string, msg interface{}) (respErr error) {
	return actorerr.ActorUnimplemented
}

func (s *Base) OnHandleEvent(event interface{}) {
	logger.KV("actorId", s.GetID()).Warn("not implement OnHandleEvent")
}
