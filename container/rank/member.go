package rank

import (
	"github.com/wwj31/dogactor/tools"
	"math"
)

type num int64 // 分数类型

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

// Score 作为 Rank.Add 第二参数，传入分数依次作为排名权重
func score(scores ...int64) []num {
	var nums []num
	for _, i64 := range scores {
		nums = append(nums, num(i64))
	}

	nums = append(nums, num(math.MaxInt64-(int64(tools.Now().Nanosecond())+_inc)))
	_inc++
	return nums
}
