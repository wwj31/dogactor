package servmesh

type MeshHandler interface {
	OnNewServ(actorId, host string, add bool)
}
