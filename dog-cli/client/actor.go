package client

import (
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/actor/cluster"
	"github.com/wwj31/dogactor/actor/cmd"
)

var sys *actor.System

func Startup(etcdAddr, etcdPrefix string) {
	sys, _ := actor.NewSystem(actor.Addr("127.0.0.1:5000"),
		cluster.WithRemote(etcdAddr, etcdPrefix),
		actor.WithCMD(cmd.New()))

	<-sys.CStop
}
