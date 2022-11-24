package ecs

import (
	"reflect"
)

type Entity struct {
	id         string
	typ        int64
	runtime    *Runtime
	components map[ComponentType]IComponent
}

func NewEntity(id string) *Entity {
	enti := &Entity{
		id:         id,
		components: make(map[ComponentType]IComponent, 0),
	}
	return enti
}

func (s *Entity) SetId(eid string) { s.id = eid }
func (s *Entity) Id() string       { return s.id }

func (s *Entity) SetType(t int64) { s.typ = t }
func (s *Entity) Type() int64     { return s.typ }

func (s *Entity) SetComponent(c ...IComponent) {
	upd := false
	for _, comp := range c {
		if reflect.ValueOf(comp).IsNil() {
			continue
		}
		if _, ok := s.components[comp.Type()]; !ok {
			upd = true
		}
		s.components[comp.Type()] = comp
	}
	if s.runtime != nil && upd {
		s.runtime.updateECS(s)
	}
}

func (s *Entity) DelComponent(c ...IComponent) {
	upd := false
	for _, comp := range c {
		if reflect.ValueOf(comp).IsNil() {
			continue
		}
		if _, ok := s.components[comp.Type()]; ok {
			upd = true
			delete(s.components, comp.Type())
		}
	}
	if s.runtime != nil && upd {
		s.runtime.updateECS(s)
	}
}

func (s *Entity) ResetComponent() {
	s.components = make(map[ComponentType]IComponent, 0)
	s.runtime.updateECS(s)
}

func (s *Entity) GetComponent(t ComponentType) IComponent {
	return s.components[t]
}

func (s *Entity) GetAllComponent() map[ComponentType]IComponent {
	return s.components
}

func (s *Entity) RangeComponent(f func(ComponentType, IComponent), ts ...ComponentType) {
	if len(ts) == 0 {
		for k, v := range s.components {
			f(k, v)
		}
	} else {
		for _, t := range ts {
			if v, ok := s.components[t]; ok {
				f(t, v)
			}
		}
	}
}

func (s *Entity) ComponentCombId() uint64 {
	var v uint64
	for _, c := range s.components {
		v |= c.Type().ComponentType()
	}
	return v
}
