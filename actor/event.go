package actor

type Ev_newActor struct {
	ActorId     string
	Publish     bool
	FromCluster bool
}

type Ev_delActor struct {
	ActorId     string
	Publish     bool
	FromCluster bool
}

type Ev_clusterUpdate struct {
	ActorId string
	Host    string
	Add     bool
}
