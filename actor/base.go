package actor

import (
	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/log"
)

type Base struct {
	Actor
}

func (s *Base) onInitActor(actor Actor) {
	s.Actor = actor
}

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
