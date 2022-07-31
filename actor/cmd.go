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

func WithCMD(cmd Cmder) SystemOption {
	return func(system *System) error {
		system.cmd = cmd
		cmd.RegistCmd("", "actorinfo", system.actorInfo, "本地actor信息")
		cmd.Start(system)
		return nil
	}
}

func (s *System) actorInfo(param ...string) {
	var actors []string

	s.actorCache.Range(func(key, value interface{}) bool {
		obj := value.(*actor)
		actors = append(actors, fmt.Sprintf("┃%-48v┃%10v     ┃%8v    ┃", key, len(obj.mailBox), cap(obj.mailBox)))
		return true
	})

	sort.SliceStable(actors, func(i, j int) bool {
		return actors[i] < actors[j]
	})

	format := `
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ local actor ━━━━━━┳━━━━━━━━━━━━━━━┳━━━━━━━━━━━━┓
┃                   actorId                      ┃ mail-box size ┃  max size  ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━╋━━━━━━━━━━━━┫
%s
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ local actor ━━━━━━┻━━━━━━━━━━━━━━━┻━━━━━━━━━━━━┛
`
	fmt.Println(fmt.Sprintf(format, strings.Join(actors, "\n")))
}
