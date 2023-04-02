package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/wwj31/dogactor/actor"
)

func TestIdleAndRunning(t *testing.T) {
	system, _ := actor.NewSystem()
	a := &actor.TmpActor{}
	_ = system.NewActor("tmp", a)
	a.Init = func() {
		fmt.Println("actor init")
	}
	a.Handle = func(msg actor.Message) {
		fmt.Println("processing msg", msg)
	}

	go func() {
		for {
			time.Sleep(65 * time.Second)
			system.Send("", a.ID(), "", "foo")
		}
	}()
	<-system.Stopped

	fmt.Println("stop")
}
