package expect

import (
	"fmt"

	"github.com/wwj31/dogactor/log"
)

func Nil(v interface{}, elseLog ...log.Fields) {
	if v != nil {
		msg := fmt.Sprintf("\n%v\n", v)
		if len(elseLog) > 0 {
			for _, v := range elseLog {
				log.KVs(v).ErrorStack(3, msg)
			}
		} else {
			log.ErrorStack(3, msg)
		}
		panic(nil)
	}
}
func True(b bool, elseLog ...log.Fields) {
	if !b {
		msg := "assert false"
		if len(elseLog) > 0 {
			for _, v := range elseLog {
				log.KVs(v).ErrorStack(3, msg)
			}
		} else {
			log.ErrorStack(3, msg)
		}
		panic(nil)
	}
}
