package actor

import (
	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/log"
)

// create actor by Base anonymously,example:
// type MyActor struct{
// 	 Base
// }

type Base struct {
	Actor
}

func (s *Base) initActor(actor Actor) {
	s.Actor = actor
}

func (s *Base) OnInit()      { log.SysLog.Warnw("actor default init", "actorId", s.ID()) }
func (s *Base) OnStop() bool { return true }

func (s *Base) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	log.SysLog.Warnw("not implement OnHandleMessage", "actorId", s.ID())
}

func (s *Base) OnHandleRequest(sourceId, targetId, requestId string, msg interface{}) (respErr error) {
	return actorerr.ActorUnimplemented
}

func (s *Base) OnHandleEvent(event interface{}) {
	log.SysLog.Warnw("not implement OnHandleEvent", "actorId", s.ID())
}

type TmpActor struct {
	Base
	Init          func()
	HandleMessage func(sourceId, targetId string, msg interface{})
	HandleRequest func(sourceId, targetId, requestId string, msg interface{}) error
	HandleEvent   func(event interface{})
}

func (s *TmpActor) OnInit() {
	if s.Init != nil {
		s.Init()
	}
}

func (s *TmpActor) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	if s.HandleMessage != nil {
		s.HandleMessage(sourceId, targetId, msg)
	}
}

func (s *TmpActor) OnHandleRequest(sourceId, targetId, requestId string, msg interface{}) (respErr error) {
	if s.HandleRequest != nil {
		return s.HandleRequest(sourceId, targetId, requestId, msg)
	}
	return nil
}

func (s *TmpActor) OnHandleEvent(event interface{}) {
	if s.HandleEvent != nil {
		s.HandleEvent(event)
	}
}
