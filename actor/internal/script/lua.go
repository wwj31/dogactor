package script

import (
	golua "github.com/yuin/gopher-lua"
	"path"

	"github.com/wwj31/dogactor/expect"
)

type ILua interface {
	Load(path string)
	Register(methodName string, f golua.LGFunction)
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
func (s *lua) Load(main string) {
	var err error
	err = s.ls.DoFile(main)
	expect.Nil(err, "main", main)

	dir, _ := path.Split(main)
	err = s.ls.CallByParam(golua.P{
		Fn:      s.ls.GetGlobal("main"),
		NRet:    0,
		Protect: true,
	}, golua.LString(dir))
	expect.Nil(err, "main", main)
}

func (s *lua) Register(methodName string, f golua.LGFunction) {
	s.ls.Register(methodName, f)
}

func (s *lua) CallFun(name string, ret int, args ...golua.LValue) []golua.LValue {
	err := s.ls.CallByParam(golua.P{
		Fn:      s.ls.GetGlobal(name),
		NRet:    ret,
		Protect: true,
	}, args...)
	expect.Nil(err, "name", name)
	var res []golua.LValue
	for i := -1; i >= -ret; i-- {
		res = append(res, s.ls.Get(i))
	}
	s.ls.Pop(ret)
	return res
}
