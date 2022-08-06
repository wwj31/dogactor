package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

type Cmd struct {
	actorId string
	usage   string
	fn      func(...string)
}

type Commands struct {
	actorSystem *actor.System
	cmds        sync.Map
}

func New() *Commands {
	cmd := &Commands{}
	cmd.RegistCmd("", "h", func(i ...string) {
		info := fmt.Sprintf("\ncommand-line 介绍:\n")
		cmd.cmds.Range(func(cmdName, cmd interface{}) bool {
			if cmdName == "h" {
				return true
			}
			info += fmt.Sprintf("%-12v usage: %v \n", cmdName, cmd.(Cmd).usage)
			return true
		})
		fmt.Println(info)
	})

	return cmd
}

func (c *Commands) Start(actorSystem *actor.System) {
	c.actorSystem = actorSystem
	go c.loop()
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

			// special windows type of ENTER
			if index := strings.Index(result, "\r"); index >= 0 {
				result = result[:index] + result[index+1:]
			}

			result = result[:len(result)-1]
			args := strings.Split(result, " ")
			cmdName := args[0]
			if cmdName == "" {
				return
			}

			ins, exist := c.cmds.Load(cmdName)
			if !exist {
				log.SysLog.Warnw("dogctl not exists", "cmdName", cmdName)
			} else {
				cmd := ins.(Cmd)
				param := args[1:]
				if cmd.actorId == "" {
					cmd.fn(param...)
				} else {
					_ = c.actorSystem.Send("", cmd.actorId, "Cmd", func() { cmd.fn(param...) })
				}
			}
		})
	}
}

func (c *Commands) RegistCmd(actorId, cmd string, fn func(...string), usage ...string) {
	var u string
	if len(usage) == 1 {
		u = usage[0]
	}
	c.cmds.LoadOrStore(cmd, Cmd{actorId: actorId, usage: u, fn: fn})
}
