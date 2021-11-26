package actor

import (
	"fmt"
	"strings"
)

type ICmd interface {
	Start(actorSystem *System)
	RegistCmd(actorId, cmd string, f func(...string), usage ...string)
}

// 设置Actor监听的端口
func WithCMD(cmd ICmd) SystemOption {
	return func(system *System) error {
		system.cmd = cmd
		cmd.RegistCmd("", "actorinfo", system.actorInfo, "本地actor信息")
		cmd.Start(system)
		return nil
	}
}

func (s *System) actorInfo(param ...string) {
	actors := []string{}
	s.actorCache.Range(func(key, value interface{}) bool {
		actors = append(actors, fmt.Sprintf("[actorId=%v mail-box=%v]", key, len(value.(*actor).mailBox)))
		return true
	})
	format := `
--------------------------------- local actor ---------------------------------
%s
--------------------------------- local actor ---------------------------------
`
	fmt.Println(fmt.Sprintf(format, strings.Join(actors, "\n")))
}
