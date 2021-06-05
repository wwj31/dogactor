package actor

import (
	"github.com/wwj31/godactor/actor/err"
)

type ActorHanlerBase struct {
	IActor
}

func (s *ActorHanlerBase) initActor(actor IActor) {
	s.IActor = actor
}

func (s *ActorHanlerBase) Init()      {}
func (s *ActorHanlerBase) Stop() bool { return true }

func (s *ActorHanlerBase) HandleMessage(sourceId, targetId string, msg interface{}) {
	logger.KV("actorId", s.GetID()).Warn("not implement HandleMessage")
}

func (s *ActorHanlerBase) HandleRequest(sourceId, targetId, requestId string, msg interface{}) (respErr error) {
	return err.ActorUnimplemented
}

func (s *ActorHanlerBase) HandleEvent(event interface{}) {
	logger.KV("actorId", s.GetID()).Warn("not implement HandleEvent")
}
func (s *ActorHanlerBase) BResponse(requestId string, msg interface{}) {
	e := s.IActor.Response(requestId, msg)
	if e != nil {
		logger.KV("err", e).ErrorStack(3, "Response error")
	}
}
