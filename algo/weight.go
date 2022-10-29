package algo

import (
	"fmt"
	"math/rand"
	"sort"
)

func WeightSelect(len int, weightFn func(index int) int) int {
	var (
		totalWeight int
		arr         = make([]int, 0, len)
	)

	for i := 0; i < len; i++ {
		val := weightFn(i)
		totalWeight += val
		arr = append(arr, totalWeight)
	}

	rd := rand.Intn(totalWeight) + 1
	idx := sort.SearchInts(arr, rd)
	if idx < len {
		return idx
	}

	panic(fmt.Sprintf("WeightSelect the index not found in arr:%v", arr))
}
