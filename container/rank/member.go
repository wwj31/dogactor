package rank

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
