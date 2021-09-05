package actor

import (
	"runtime"
	"time"

	"github.com/wwj31/dogactor/expect"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
	lua "github.com/yuin/gopher-lua"
)

func (s *actor) CallLua(name string, ret int, args ...lua.LValue) []lua.LValue {
	if s.lua == nil {
		log.WarnStack(4, "call lua, the actor was not Call actor.SetLua()")
		return nil
	}
	return s.lua.CallFun(name, ret, args...)
}

func (s *actor) register2Lua() {
	s.lua.Register("govn", LGoVersion)
	s.lua.Register("debug", LDebug)
	s.lua.Register("warn", LWarn)
	s.lua.Register("error", LError)
	s.lua.Register("addtimer", s.LAddTimer)
	s.lua.Register("canceltimer", s.LCancelTimer)
}

func LGoVersion(l *lua.LState) int {
	v := runtime.Version()
	l.Push(lua.LString(v))
	return 1
}

func LDebug(l *lua.LState) int {
	str := l.ToString(-1)
	l.Pop(1)
	log.Debug(str)
	return 0
}
func LWarn(l *lua.LState) int {
	str := l.ToString(-1)
	l.Pop(1)
	log.Warn(str)
	return 0
}
func LError(l *lua.LState) int {
	str := l.ToString(-1)
	l.Pop(1)
	log.Error(str)
	return 0
}

func (s *actor) LAddTimer(l *lua.LState) int {
	interval := l.ToInt64(-1)
	count := l.ToInt(-2)
	callback := l.ToFunction(-3)
	l.Pop(3)

	if count == 0 || interval < 0 {
		log.KVs(log.Fields{"count": count, "timeout": interval}).Error("val error")
		return 0
	}

	id := s.AddTimer(tools.UUID(), time.Duration(interval), func(dt int64) {
		err := l.CallByParam(lua.P{
			Fn:      callback,
			NRet:    0,
			Protect: true,
		})
		expect.Nil(err)
	}, int32(count))
	l.Push(lua.LString(id))
	return 1
}

func (s *actor) LCancelTimer(l *lua.LState) int {
	timerId := l.ToString(-1)
	l.Pop(1)
	s.CancelTimer(timerId)
	return 0
}
