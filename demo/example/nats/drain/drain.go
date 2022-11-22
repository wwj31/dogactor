package main

import (
	"fmt"
	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/actor/cluster/mq"
	"github.com/wwj31/dogactor/actor/cluster/mq/nats"
	"github.com/wwj31/dogactor/demo/example/nats/msg"
	"github.com/wwj31/dogactor/logger"
	"github.com/wwj31/dogactor/tools"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var natsUrl = "nats://localhost:4222"

func main() {
	protoIndex := tools.NewProtoIndex(func(name string) (interface{}, bool) {
		return msg.Spawner(name)
	}, tools.EnumIdx{})

	system1, _ := actor.NewSystem(mq.WithRemote(natsUrl, nats.New()), actor.ProtoIndex(protoIndex))
	system2, _ := actor.NewSystem(mq.WithRemote(natsUrl, nats.New()), actor.ProtoIndex(protoIndex))

	name := "drainActor"
	drainActor := actor.TmpActor{}
	drainActor.HandleMessage = func(sourceId, targetId actor.Id, v interface{}) {
		drainData := v.(*msg.DrainTest)
		fmt.Println("receive msg:", drainData.Data)
		time.Sleep(time.Duration(rand.Intn(100)+100) * time.Millisecond)
	}

	_ = system1.Add(actor.New(name, &drainActor, actor.SetMailBoxSize(50)))

	go func() {
		for {
			_ = system2.Send("", name, "", &msg.DrainTest{Data: time.Now().String()})
			time.Sleep(200 * time.Millisecond)
		}
	}()

	time.Sleep(20 * time.Second)
	fmt.Println("draining")
	drainActor.Drain(func() {
		fmt.Println("drain over")
	})

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-c
	system1.Stop()
	system2.Stop()
	<-system1.Stopped
	<-system2.Stopped

	logger.Close()
}
