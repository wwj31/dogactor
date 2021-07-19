package actor

import (
	"github.com/wwj31/godactor/actor/err"
	"github.com/wwj31/godactor/log"
)

type Base struct {
	IActor
}

func (s *Base) onInitActor(actor IActor) {
	s.IActor = actor
}

func (s *Base) OnInit()      { log.KV("actor", s.GetID()).Warn("actor default init") }
func (s *Base) OnStop() bool { return true }

func (s *Base) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	logger.KV("actorId", s.GetID()).Warn("not implement OnHandleMessage")
}

func (s *Base) OnHandleRequest(sourceId, targetId, requestId string, msg interface{}) (respErr error) {
	return err.ActorUnimplemented
}

func (s *Base) OnHandleEvent(event interface{}) {
	logger.KV("actorId", s.GetID()).Warn("not implement OnHandleEvent")
}
