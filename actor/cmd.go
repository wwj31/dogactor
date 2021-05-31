package actor

import (
	"fmt"
	"strings"
)

type ICmd interface {
	Start(actorSystem *ActorSystem)
	RegistCmd(actorId, cmd string, f func(...string))
}

// 设置Actor监听的端口
func WithCMD(cmd ICmd) SystemOption {
	return func(system *ActorSystem) error {
		system.cmd = cmd
		cmd.RegistCmd("", "actorinfo", system.actorInfo)
		cmd.RegistCmd("", "loadlua", system.loadlua)
		cmd.Start(system)
		return nil
	}
}

func (as *ActorSystem) RegistCmd(actorId, cmd string, fn func(...string)) {
	if as.cmd != nil {
		as.cmd.RegistCmd(actorId, cmd, fn)
	}
}

func (s *ActorSystem) actorInfo(param ...string) {
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
	logger.Info(fmt.Sprintf(format, strings.Join(actors, "\n")))
}

func (s *ActorSystem) loadlua(param ...string) {
	s.actorCache.Range(func(key, value interface{}) bool {
		a := value.(*actor)
		if a.lua != nil {
			s.Send("", a.id, "LOAD LUA", func() { a.lua.Load(a.luapath) })
		}
		return true
	})
}
