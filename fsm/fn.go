package fsm

type StateFnHandler struct {
	StateFn  int
	EnterFn  func(fsm *FSM)
	LeaveFn  func(fsm *FSM)
	HandleFn func(fsm *FSM, v ...interface{})
}

func (s *StateFnHandler) State() int { return s.StateFn }

func (s *StateFnHandler) Enter(fsm *FSM)                    { s.EnterFn(fsm) }
func (s *StateFnHandler) Leave(fsm *FSM)                    { s.LeaveFn(fsm) }
func (s *StateFnHandler) Handle(fsm *FSM, v ...interface{}) { s.HandleFn(fsm, v...) }
