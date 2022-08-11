package fsm

import (
	"testing"
	"time"
)

const (
	A int = iota
	B
	C
	D
)

func TestName(t *testing.T) {
	stateABCD := New()
	stateABCD.Add(&StateA{property: "I'm State A"})
	stateABCD.Add(&StateB{property: "I'm State B"})
	stateABCD.Add(&StateC{property: "I'm State C"})
	stateABCD.Add(&StateD{
		property: "I'm State D",
		StateFnHandler: StateFnHandler{
			StateFn: D,
			EnterFn: func(fsm *FSM) {
				println("enter D")
			},
			LeaveFn: func(fsm *FSM) {
				println("Leave D")
			},
			HandleFn: func(fsm *FSM, v ...interface{}) {
				println("Handle D switch->A")
				fsm.Switch(A)
			},
		},
	})

	stateABCD.Switch(A)
	timer := time.NewTicker(1 * time.Second)

	for range timer.C {
		stateABCD.Handle(nil)
	}
}

// A
type StateA struct {
	property string
}

func (s *StateA) State() int                        { return A }
func (s *StateA) Enter(fsm *FSM)                    { println("enter A"); println(s.propertyA) }
func (s *StateA) Leave(fsm *FSM)                    { println("Leave A") }
func (s *StateA) Handle(fsm *FSM, v ...interface{}) { println("Handle A switch->B"); fsm.Switch(B) }

// B
type StateB struct {
	property string
}

func (s *StateB) State() int                        { return B }
func (s *StateB) Enter(fsm *FSM)                    { println("enter B") }
func (s *StateB) Leave(fsm *FSM)                    { println("Leave B") }
func (s *StateB) Handle(fsm *FSM, v ...interface{}) { println("Handle B switch->C"); fsm.Switch(C) }

// C
type StateC struct {
	property string
}

func (s *StateC) State() int                        { return C }
func (s *StateC) Enter(fsm *FSM)                    { println("enter C") }
func (s *StateC) Leave(fsm *FSM)                    { println("Leave C") }
func (s *StateC) Handle(fsm *FSM, v ...interface{}) { println("Handle C switch->D"); fsm.Switch(D) }

// C
type StateD struct {
	property string
	StateFnHandler
}
