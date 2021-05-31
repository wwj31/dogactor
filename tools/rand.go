package tools

import (
	"math/rand"

	"github.com/wwj31/godactor/log"
)

// 左闭右开 [x,y)
func Randx_y(x, y int) int {
	if x > y {
		log.KVs(log.Fields{"x": x, "y": y}).ErrorStack(3, "x>y;")
		return 0
	} else if x == y {
		return x
	}
	return rand.Intn(y-x) + x
}

// 简单的概率ratio 1-100
func Probability(ratio int) bool {
	v := Randx_y(0, 100)
	return v < ratio
}

// 简单的概率ratio 1-10000
func Probability10000(ratio int) bool {
	v := Randx_y(0, 10000)
	return v < ratio
}
