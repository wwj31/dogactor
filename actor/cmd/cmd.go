package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/wwj31/godactor/actor"
	"github.com/wwj31/godactor/log"
	"github.com/wwj31/godactor/tools"
)

type Cmd struct {
	actorId string
	fn      func(...string)
}

type Commands struct {
	actorSystem *actor.ActorSystem
	cmds        sync.Map
}

func New() *Commands {
	cmd := &Commands{}
	cmd.RegistCmd("", "h", func(i ...string) {
		info := fmt.Sprintf("\ninstructions:\n")
		cmd.cmds.Range(func(cmdName, cmd interface{}) bool {
			info += fmt.Sprintf("%v \n", cmdName)
			return true
		})
		log.Info(info)
	})

	return cmd
}

func (c *Commands) Start(actorSystem *actor.ActorSystem) {
	c.actorSystem = actorSystem
	tools.GoEngine(c.loop)
}

func (c *Commands) loop() {
	for {
		tools.Try(func() {
			reader := bufio.NewReader(os.Stdin)
			result, err := reader.ReadString('\n')
			if err != nil || len(result) == 0 {
				fmt.Println("read error:", err)
				return
			}

			result = result[:len(result)-1]
			args := strings.Split(result, " ")
			cmdName := args[0]
			if cmdName == "" {
				return
			}

			ins, exist := c.cmds.Load(cmdName)
			if !exist {
				log.KV("cmdName", cmdName).Info("cmd not exists")
			} else {
				cmd := ins.(Cmd)
				param := args[1:]
				if cmd.actorId == "" {
					cmd.fn(param...)
				} else {
					c.actorSystem.Send("", cmd.actorId, "Cmd", func() { cmd.fn(param...) })
				}
			}
		}, nil)
	}
}

func (c *Commands) RegistCmd(actorId, cmd string, fn func(...string)) {
	c.cmds.LoadOrStore(cmd, Cmd{actorId: actorId, fn: fn})
}
