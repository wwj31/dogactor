package servmesh

type ServMeshHander interface {
	OnNewServ(actorId, host string, add bool)
}
