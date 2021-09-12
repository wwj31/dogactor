package servmesh_provider

type EtcdHander interface {
	OnEtcdNew(k, v string)
}
