package servmesh_provider

type ServMeshHander interface {
	OnNewServ(actorId, host string, add bool)
}
