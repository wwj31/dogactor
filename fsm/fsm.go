package fsm

import (
	"fmt"
)

type StateHandler interface {
	State() int
	Enter(*FSM)
	Leave(*FSM)
	Handle(*FSM, ...interface{})
}

type (
	FSM struct {
		currState int                  // 当前状态
		states    map[int]StateHandler // 状态处理对象
	}
)

func New() *FSM {
	fsm := &FSM{}
	fsm.currState = -1
	fsm.states = make(map[int]StateHandler, 2)
	return fsm
}

func (s *FSM) Add(state StateHandler) error {
	if _, exit := s.states[state.State()]; exit {
		return fmt.Errorf("repeated state:%v", state.State())
	}
	s.states[state.State()] = state
	return nil
}

func (s *FSM) State() int {
	return s.currState
}

func (s *FSM) ForceState(state int) {
	s.currState = state
}

func (s *FSM) CurrentStateHandler() StateHandler {
	return s.states[s.currState]
}

func (s *FSM) Switch(nextState int) error {
	if _, exit := s.states[nextState]; !exit {
		return fmt.Errorf("not found next_state:%v", nextState)
	}

	if s.State() != -1 {
		s.states[s.State()].Leave(s)
	}
	s.currState = nextState
	s.states[s.State()].Enter(s)

	return nil
}

func (s *FSM) Handle(v interface{}) {
	s.CurrentStateHandler().Handle(s, v)
}
