package actor

import (
	"github.com/wwj31/godactor/actor/err"
	"github.com/wwj31/godactor/log"
)

type HandlerBase struct {
	IActor
}

func (s *HandlerBase) onInitActor(actor IActor) {
	s.IActor = actor
}

func (s *HandlerBase) OnInit()      { log.KV("actor", s.GetID()).Warn("actor default init") }
func (s *HandlerBase) OnStop() bool { return true }

func (s *HandlerBase) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	logger.KV("actorId", s.GetID()).Warn("not implement OnHandleMessage")
}

func (s *HandlerBase) OnHandleRequest(sourceId, targetId, requestId string, msg interface{}) (respErr error) {
	return err.ActorUnimplemented
}

func (s *HandlerBase) OnHandleEvent(event interface{}) {
	logger.KV("actorId", s.GetID()).Warn("not implement OnHandleEvent")
}
