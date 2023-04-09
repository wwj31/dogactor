package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/wwj31/dogactor/actor"
)

func TestIdleAndRunning(t *testing.T) {
	system, _ := actor.NewSystem()
	a := &actor.BaseFunc{}
	_ = system.NewActor("actor1", a)
	a.Init = func() {
		fmt.Println("actor init")
	}
	a.Handle = func(msg actor.Message) {
		fmt.Println("processing msg ", msg.RawMsg())
	}

	go func() {
		for {
			time.Sleep(10 * time.Second)
			system.Send("", a.ID(), "", "foo")
		}
	}()
	<-system.Stopped

	fmt.Println("stop")
}
