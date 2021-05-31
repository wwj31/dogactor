package ecs

import (
	"errors"
	"fmt"
)

type EntityCollection struct {
	entities map[string]*Entity
}

func newEntityCollection() *EntityCollection {
	return &EntityCollection{
		entities: make(map[string]*Entity),
	}
}

func (s *EntityCollection) add(ent *Entity) error {
	if ent == nil {
		return errors.New("ent == nil")
	}
	if _, ok := s.entities[ent.Id()]; ok {
		return errors.New(fmt.Sprint("repeated eid:%v", ent.Id()))
	}
	s.entities[ent.Id()] = ent
	return nil
}

func (s *EntityCollection) get(eid string) *Entity {
	return s.entities[eid]
}
func (s *EntityCollection) all() map[string]*Entity {
	return s.entities
}

func (s *EntityCollection) del(eid string) bool {
	if _, ok := s.entities[eid]; !ok {
		return false
	}
	delete(s.entities, eid)
	return true
}
