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
