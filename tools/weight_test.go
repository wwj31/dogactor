package tools

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestWeight(t *testing.T) {
	arr := [][]int{
		{33, 0},
		{33, 1},
		{33, 3},
	}

	count := map[int]int{}
	rand.Seed(time.Now().UnixMicro())
	for i := 0; i < 100000; i++ {
		idx := WeightSelect(len(arr), func(index int) int {
			return arr[index][0]
		})
		count[idx] = count[idx] + 1
	}
	fmt.Println(count)
}
