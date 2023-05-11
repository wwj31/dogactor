package event

type EvNewActor struct {
	ActorId     string
	FromCluster bool
}

type EvDelActor struct {
	ActorId     string
	Publish     bool
	FromCluster bool
}

type EvActorSubMqFin struct {
	ActorId string
}
