package actor

import "github.com/wwj31/dogactor/log"

// create actor by Base anonymously,example:
// type MyActor struct{
// 	 Base
// }

type Base struct {
	Actor
	Timer
	Messenger
	Drainer
}

func (s *Base) Exit() {
	s.Timer.clear()
	s.Actor.Exit()
}

func (s *Base) OnInit()              { log.SysLog.Warnw("actor default init", "actorId", s.ID()) }
func (s *Base) OnHandle(msg Message) { log.SysLog.Errorw("actor default hande", "actorId", s.ID()) }
func (s *Base) OnStop() bool         { return true }

func (s *Base) initActor(actor Actor) {
	s.Actor = actor
	s.Timer = newTimerScheduler(s)
	s.Messenger = newCommunication(s)
	s.Drainer = newDrain(s)
}

// create functional actor by BaseFunc ,example:
// bar := BaseFunc{}
// bar.Init = func(){}
// bar.Handle = func(msg Message){}

type BaseFunc struct {
	Base
	Init   func()
	Stop   func() bool
	Handle func(msg Message)
}

func (s *BaseFunc) OnInit() {
	if s.Init != nil {
		s.Init()
	}
}

func (s *BaseFunc) OnStop() bool {
	if s.Stop != nil {
		return s.Stop()
	}

	return true
}

func (s *BaseFunc) OnHandle(msg Message) {
	if s.Handle != nil {
		s.Handle(msg)
	}
}
