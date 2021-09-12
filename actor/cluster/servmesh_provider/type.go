package servmesh_provider

type ServMeshHander interface {
	OnNewServ(k, v string)
}
