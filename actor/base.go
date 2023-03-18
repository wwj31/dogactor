package actor

import (
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

func (s *Base) OnInit()              { log.SysLog.Warnw("actor default init", "actorId", s.ID()) }
func (s *Base) OnStop() bool         { return true }
func (s *Base) OnHandle(msg Message) { log.SysLog.Warnw("actor default hande", "actorId", s.ID()) }

// create functional actor by TmpActor ,example:
// bar := TmpActor{}
// bar.Init = func(){}
// bar.Handle = func(msg Message){}

type TmpActor struct {
	Base
	Init   func()
	Stop   func() bool
	Handle func(msg Message)
}

func (s *TmpActor) OnInit() {
	if s.Init != nil {
		s.Init()
	}
}

func (s *TmpActor) OnStop() bool {
	if s.Stop != nil {
		return s.Stop()
	}

	return true
}

func (s *TmpActor) OnHandle(msg Message) {
	if s.Handle != nil {
		s.Handle(msg)
	}
}
