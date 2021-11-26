package actor

type EvNewactor struct {
	ActorId     string
	Publish     bool
	FromCluster bool
}

type EvDelactor struct {
	ActorId     string
	Publish     bool
	FromCluster bool
}

type EvClusterupdate struct {
	ActorId string
	Host    string
	Add     bool
}

type EvSessionclosed struct {
	PeerHost string
}
type EvSessionopened struct {
	PeerHost string
}
