package expect

import (
	"fmt"
	"github.com/wwj31/dogactor/l"
)

func Nil(v interface{}, args ...interface{}) {
	if v != nil {
		msg := fmt.Sprintf("\n%v\n", v)
		l.Errorw(msg,args...)
		panic(nil)
	}
}
func True(b bool, args ...interface{}) {
	if !b {
		msg := "assert false"
		l.Errorw(msg,args...)
		panic(nil)
	}
}
