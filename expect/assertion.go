package expect

import (
	"fmt"
)

func Nil(v interface{}, args ...interface{}) {
	if v != nil {
		msg := fmt.Sprintf("\n%v %v\n", v, fmt.Sprint(args...))
		panic(msg)
	}
}
