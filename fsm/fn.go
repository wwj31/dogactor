package fsm

type StateFnHandler struct {
	StateFn  int
	EnterFn  func(fsm *FSM)
	LeaveFn  func(fsm *FSM)
	HandleFn func(fsm *FSM, v ...interface{})
}

func (s *StateFnHandler) State() int { return s.StateFn }

func (s *StateFnHandler) Enter(fsm *FSM) {
	if s.EnterFn != nil {
		s.EnterFn(fsm)
	}
}
func (s *StateFnHandler) Leave(fsm *FSM) {
	if s.LeaveFn != nil {
		s.LeaveFn(fsm)
	}
}
func (s *StateFnHandler) Handle(fsm *FSM, v ...interface{}) {
	if s.HandleFn != nil {
		s.HandleFn(fsm, v...)
	}
}
