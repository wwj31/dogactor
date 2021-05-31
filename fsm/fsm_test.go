package fsm

import (
	"testing"
	"time"
)

const (
	A int = iota
	B
	C
)

func TestName(t *testing.T) {
	stateABC := New()
	stateABC.Add(&StateA{propertyA: "I'm State A"})
	stateABC.Add(&StateB{propertyB: "I'm State B"})
	stateABC.Add(&StateC{propertyC: "I'm State C"})

	stateABC.Switch(A)
	timer := time.NewTicker(1 * time.Second)

	for range timer.C {
		stateABC.Handle(nil)
	}
}

// A
type StateA struct {
	propertyA string
}

func (s *StateA) State() int                     { return A }
func (s *StateA) Enter(fsm *FSM)                 { println("enter A") }
func (s *StateA) Leave(fsm *FSM)                 { println("Leave A") }
func (s *StateA) Handle(fsm *FSM, v interface{}) { println("Handle A swich->B"); fsm.Switch(B) }

// B
type StateB struct {
	propertyB string
}

func (s *StateB) State() int                     { return B }
func (s *StateB) Enter(fsm *FSM)                 { println("enter B") }
func (s *StateB) Leave(fsm *FSM)                 { println("Leave B") }
func (s *StateB) Handle(fsm *FSM, v interface{}) { println("Handle B swich->C"); fsm.Switch(C) }

// C
type StateC struct {
	propertyC string
}

func (s *StateC) State() int                     { return C }
func (s *StateC) Enter(fsm *FSM)                 { println("enter C") }
func (s *StateC) Leave(fsm *FSM)                 { println("Leave C") }
func (s *StateC) Handle(fsm *FSM, v interface{}) { println("Handle C swich->A"); fsm.Switch(A) }
