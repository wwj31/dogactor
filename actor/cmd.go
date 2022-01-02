package actor

import (
	"fmt"
	"sort"
	"strings"
)

type Cmder interface {
	Start(actorSystem *System)
	RegistCmd(actorId, cmd string, f func(...string), usage ...string)
}

// 设置Actor监听的端口
func WithCMD(cmd Cmder) SystemOption {
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
		actors = append(actors, fmt.Sprintf("┃%-46v┃%10v%6v", key, len(value.(*actor).mailBox), "┃"))
		return true
	})
	sort.SliceStable(actors, func(i, j int) bool {
		return actors[i] < actors[j]
	})
	format := `
┏━━━━━━━━━━━━━━━━━━━━━━━━━━local actor━━━━━━━━━┳━━━━━━━━━━━━━━━┓
┃                   actorId                    ┃ mail-box size ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━┫
%s
┗━━━━━━━━━━━━━━━━━━━━━━━━━━local actor━━━━━━━━━┻━━━━━━━━━━━━━━━┛
`
	fmt.Println(fmt.Sprintf(format, strings.Join(actors, "\n")))
}
