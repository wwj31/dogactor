package actor

type EvNewActor struct {
	ActorId     Id
	Publish     bool
	FromCluster bool
}

type EvDelActor struct {
	ActorId     Id
	Publish     bool
	FromCluster bool
}

type EvClusterUpdate struct {
	ActorId Id
	Host    string
	Add     bool
}

type EvSessionClosed struct {
	PeerHost string
}

type EvSessionOpened struct {
	PeerHost string
}

type ReqMsgDrain struct{}
type RespMsgDrain struct{ Err error }
