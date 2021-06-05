<p align="center">game server framework</a></p>

a go framework base on actor model.
it has implemented server discovery using ETCD.

Getting Started
===============

# Installing
#### To start using godactor, install Go and run `go get`:
```sh
$ go get -u github.com/wwj31/godactor
```

```go
package main

import "github.com/wwj31/godactor"

func main() {
	actorSystem1, _ := actor.System(
		actor.Addr("127.0.0.1:1111"),
		actor.WithCMD(cmd.New()),
		actor.WithEvent(event.NewActorEvent()),
		actor.WithCluster(cluster.NewCluster(etcd.NewEtcd("127.0.0.1:2379", "demo/"), remote_planc.NewRemoteMgr())),
	)
}
```