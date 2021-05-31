package ecs

type IComponent interface {
	Type() ComponentType
	EId() string
	Init(EId string)
}

type ComponentType interface {
	ComponentType() uint64
	String() string
}

type ComponentBase struct {
	eid string
}

func (s *ComponentBase) EId() string {
	return s.eid
}
func (s *ComponentBase) Init(EId string) {
	s.eid = EId
}
