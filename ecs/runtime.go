package ecs

import (
	"errors"
	"fmt"
	"github.com/wwj31/jtimer"
)

type Runtime struct {
	// all systems
	systems []ISystem
	// all entities
	entities *EntityCollection
	// all Singleton Component
	singleComponents map[ComponentType]IComponent
}

func NewRuntime() *Runtime {
	rt := &Runtime{
		systems:          make([]ISystem, 0),
		entities:         newEntityCollection(),
		singleComponents: make(map[ComponentType]IComponent),
	}
	return rt
}

// 返回此次更新运行总时长
func (s *Runtime) Run(dt float64, time_set map[string]int64) int64 {
	var t int64
	for _, sys := range s.systems {
		st1 := jtimer.Milliseconds()

		sys.UpdateFrame(dt)

		st2 := jtimer.Milliseconds()
		st := st2 - st1
		time_set[sys.base().Type()] = st
		t += st
	}
	return t
}

//entity operate : add
func (s *Runtime) AddEntity(ent *Entity) error {
	ent.runtime = s
	s.updateECS(ent)
	return s.entities.add(ent)
}

//entity operate : delete
func (s *Runtime) DeleteEntity(eid string) bool {
	for _, sys := range s.systems {
		sys.base().delTuple(eid)
	}
	return s.entities.del(eid)
}

//entity 组件发生变化，更新system关心的组件元组
func (s *Runtime) updateECS(ent *Entity) {
	combId := ent.ComponentCombId()
	for _, sys := range s.systems {
		scomp := sys.EssentialComp()
		if sys.base().compTuples[ent.Id()] == nil && combId&scomp == scomp {
			sys.base().newTuple(ent.Id(), ent.GetAllComponent())
		} else if sys.base().compTuples[ent.Id()] != nil && combId&scomp != scomp {
			sys.base().delTuple(ent.Id())
		}
	}
}

func (s *Runtime) AddSingleComponent(components ...IComponent) error {
	for _, c := range components {
		for _, comp := range s.singleComponents {
			if c.Type() == comp.Type() {
				return errors.New(fmt.Sprintf("Singleton Component type repeated:%v", c.Type()))
			}
		}
	}
	for _, c := range components {
		s.singleComponents[c.Type()] = c
	}
	return nil
}
func (s *Runtime) SingleComponent(componentType ComponentType) IComponent {
	return s.singleComponents[componentType]
}

func (s *Runtime) GetEntity(eid string) *Entity {
	return s.entities.get(eid)
}

func (s *Runtime) AddSystem(news ...ISystem) error {
	for _, n := range news {
		for _, sys := range s.systems {
			if sys.base().Type() == n.base().Type() {
				return errors.New(fmt.Sprintf("systems type repeated:%v", n.base().Type()))
			}
		}
	}
	s.systems = append(s.systems, news...)
	return nil
}

func (s *Runtime) RangeSystem(f func(ISystem) bool) {
	for _, sys := range s.systems {
		if !f(sys) {
			return
		}
	}
}
