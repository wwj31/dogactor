package tools

import (
	"math/rand"

	"github.com/wwj31/dogactor/log"
)

type CommonRand struct {
	Id     int
	Chance int32
	Extra  int32
}

func CircleRand(commonArrs []CommonRand) (id int, arrIndex int) {
	arr := make([]CommonRand, len(commonArrs))
	copy(arr, commonArrs)
	var sum int32
	for _, v := range arr {
		sum += v.Chance
	}
	var base int32
	if sum > 0 {
		chance := rand.Intn(int(sum + 1))
		for index, v := range arr {
			if v.Chance == 0 {
				continue
			}
			if int(base+v.Chance) >= chance {
				return v.Id, index
			} else {
				base += v.Chance
			}
		}
	}
	log.KV("arr", arr).ErrorStack(3, "CircleRand Wrong")
	return 0, -1
}
func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}
func Int32Merge(h, l int16) (id int32) {
	return int32(uint32(h)<<16 | uint32(l))
}

func Int32Split(id int32) (h, l int16) {
	return int16(uint32(id) >> 16), int16(uint16(id))
}
