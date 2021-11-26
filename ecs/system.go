package ecs

import "reflect"

type System interface {
	UpdateFrame(dt float64)
	EssentialComp() uint64
	base() *SystemBase
}

type SystemBase struct {
	typ        string
	runtime    *Runtime
	compTuples map[string]ITuple // EID=>ITuple
	tupleType  reflect.Type
}

func NewSystemBase(t string, rt *Runtime, tt reflect.Type) *SystemBase {
	return &SystemBase{
		typ:        t,
		runtime:    rt,
		tupleType:  tt,
		compTuples: make(map[string]ITuple),
	}
}

func (s *SystemBase) base() *SystemBase { return s }

func (s *SystemBase) Type() string      { return s.typ }
func (s *SystemBase) Runtime() *Runtime { return s.runtime }

func (s *SystemBase) Range(f func(eid string, tuple ITuple) bool) {
	for id, e := range s.compTuples {
		if !f(id, e) {
			return
		}
	}
}
func (s *SystemBase) GetTuple(EId string) ITuple {
	return s.compTuples[EId]
}
func (s *SystemBase) newTuple(EId string, comps map[ComponentType]IComponent) {
	if s.tupleType != nil {
		tuple := reflect.New(s.tupleType.Elem()).Interface().(ITuple)
		tuple.Init(comps)
		s.compTuples[EId] = tuple
	}
}

func (s *SystemBase) delTuple(EId string) {
	delete(s.compTuples, EId)
}
