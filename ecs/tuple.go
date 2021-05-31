package ecs

type ITuple interface {
	Init(comps map[ComponentType]IComponent)
}
