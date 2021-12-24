package tools

import (
	"fmt"
	"github.com/wwj31/dogactor/log"
	"runtime/debug"
)

func Try(fn func(), catch ...func(ex interface{})) {
	defer func() {
		if r := recover(); r != nil {
			stack := fmt.Sprintf("panic recover:%v\n%v", r, string(debug.Stack()))
			log.SysLog.Errorf(stack)
			if len(catch) > 0 {
				catch[0](r)
			}
		}
	}()
	fn()
}

//特别说明 于$GOROOT/src/runtime/proc.go中加入此函数
//func GoroutineId() int64 {
//  _g_ := getg()
//  return _g_.goid
//}

//func GoroutineId() uint64 {
//	b := make([]byte, 64)
//	b = b[:runtime.Stack(b, false)]
//	b = bytes.TrimPrefix(b, []byte("goroutine "))
//	b = b[:bytes.IndexByte(b, ' ')]
//	n, _ := strconv.ParseUint(string(b), 10, 64)
//	return n
//}
