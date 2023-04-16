package actor

import "github.com/wwj31/dogactor/log"

// create the actor by Base anonymously,for example:
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

func (s *Base) OnInit()          { log.SysLog.Warnw("must override OnInit", "actorId", s.ID()) }
func (s *Base) OnHandle(Message) { log.SysLog.Errorw("must override OnHandle", "actorId", s.ID()) }
func (s *Base) OnStop() bool     { return true }

func (s *Base) initActor(actor *actor) {
	s.Actor = actor
	s.Timer = newTimerScheduler(s)
	s.Messenger = newCommunication(s)
	s.Drainer = newDrain(s)
}

// create the functional actor by BaseFunc ,for example:
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
