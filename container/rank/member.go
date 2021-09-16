package rank

import (
	"math"
	"time"
)

type Key string

type Member struct {
	Key    string
	Scores []num
}

func (s Member) Less(other interface{}) bool {
	min := len(s.Scores)
	omember := other.(Member)
	if min > len(omember.Scores) {
		min = len(omember.Scores)
	}
	for i := 0; i < min; i++ {
		if s.Scores[i] > omember.Scores[i] {
			return true
		} else if s.Scores[i] < omember.Scores[i] {
			return false
		}
	}
	return len(omember.Scores) < len(s.Scores)
}

var _inc int64

// score相同,排名依次对比scores大小
func Score(scores ...int64) []num {
	nums := []num{}
	for _, i64 := range scores{
		nums = append(nums, num(i64))
	}

	nums = append(nums, num(math.MaxInt64-(int64(time.Now().Nanosecond())+_inc)))
	_inc++
	return nums
}
