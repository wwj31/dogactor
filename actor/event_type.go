package actor

type EvNewActor struct {
	ActorId     string
	Publish     bool
	FromCluster bool
}

type EvDelActor struct {
	ActorId     string
	Publish     bool
	FromCluster bool
}

type EvClusterUpdate struct {
	ActorId string
	Host    string
	Add     bool
}

type EvSessionClosed struct {
	PeerHost string
}
type EvSessionOpened struct {
	PeerHost string
}
