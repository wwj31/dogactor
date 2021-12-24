package actor

import (
	"github.com/wwj31/dogactor/actor/actorerr"
	"github.com/wwj31/dogactor/actor/log"
)

// create actor by Anonymous Base,example:
// type MyActor struct{
// 	 Base
// }

type Base struct {
	Actor
}

func (s *Base) initActor(actor Actor) {
	s.Actor = actor
}

func (s *Base) OnInit()      { log.SysLog.Warnw("actor default init","actorId",s.ID()) }
func (s *Base) OnStop() bool { return true }

func (s *Base) OnHandleMessage(sourceId, targetId string, msg interface{}) {
	log.SysLog.Warnw("not implement OnHandleMessage","actorId",s.ID())
}

func (s *Base) OnHandleRequest(sourceId, targetId, requestId string, msg interface{}) (respErr error) {
	return actorerr.ActorUnimplemented
}

func (s *Base) OnHandleEvent(event interface{}) {
	log.SysLog.Warnw("not implement OnHandleEvent","actorId",s.ID())
}
