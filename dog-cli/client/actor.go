package client

import (
	"fmt"
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/actor/cluster"
	"github.com/wwj31/dogactor/actor/cmd"
	"time"
)

type Cli struct {
	actor.Base
}

func (c *Cli) OnInit() {}

func Startup(etcdAddr, etcdPrefix string) {
	sys, _ := actor.NewSystem(actor.Addr("127.0.0.1:8760"),
		cluster.WithRemote(etcdAddr, etcdPrefix),
		actor.WithCMD(cmd.New()))

	cli := &Cli{}
	_ = sys.Add(actor.New("cli", cli))

	time.Sleep(1 * time.Second)
	rsp, err := cli.RequestWait(sys.Cluster().ID(), "nodeinfo")
	if err != nil {
		fmt.Println("RequestWait err", err)
		return
	}
	data := rsp.(string)
	fmt.Println(data)

	<-sys.CStop
}
