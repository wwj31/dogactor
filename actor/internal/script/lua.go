package script

import (
	golua "github.com/yuin/gopher-lua"
	"path"

	"github.com/wwj31/dogactor/expect"
	"github.com/wwj31/dogactor/log"
)

type ILua interface {
	Load(path string)
	Register(fname string, f golua.LGFunction)
	CallFun(name string, ret int, args ...golua.LValue) []golua.LValue
}

type lua struct {
	ls *golua.LState
}

func New() ILua {
	l := &lua{}
	state := golua.NewState()
	state.OpenLibs()
	l.ls = state
	return l
}
func (s *lua) Load(lmain string) {
	var err error
	err = s.ls.DoFile(lmain)
	expect.Nil(err, log.Fields{"lmain": lmain})

	dir, _ := path.Split(lmain)
	err = s.ls.CallByParam(golua.P{
		Fn:      s.ls.GetGlobal("main"),
		NRet:    0,
		Protect: true,
	}, golua.LString(dir))
	expect.Nil(err, log.Fields{"lmain": lmain})
}

func (s *lua) Register(fname string, f golua.LGFunction) {
	s.ls.Register(fname, f)
}

func (s *lua) CallFun(name string, ret int, args ...golua.LValue) []golua.LValue {
	err := s.ls.CallByParam(golua.P{
		Fn:      s.ls.GetGlobal(name),
		NRet:    ret,
		Protect: true,
	}, args...)
	expect.Nil(err, log.Fields{"name": name})
	res := []golua.LValue{}
	for i := -1; i >= -ret; i-- {
		res = append(res, s.ls.Get(i))
	}
	s.ls.Pop(ret)
	return res
}
