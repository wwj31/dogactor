package rank

type Score = int64 // 分数类型

type Member struct {
	ID     string
	Scores []Score
}

// Less 对m和other进行比较，返回m是否应该排在other的前面，含义同golang中sort.Interface的Less
func (m Member) Less(other interface{}) bool {
	l := len(m.Scores)
	member := other.(Member)
	if l > len(member.Scores) {
		l = len(member.Scores)
	}

	// 按照优先级先比较分数
	for i := 0; i < l-1; i++ {
		if m.Scores[i] != member.Scores[i] {
			return m.Scores[i] > member.Scores[i]
		}
	}

	// 分数相同，长度一致比插入先后
	if len(m.Scores) == len(member.Scores) {
		t1 := m.Scores[len(m.Scores)-1]
		t2 := member.Scores[len(member.Scores)-1]
		if t1 != t2 {
			return t1 < t2
		}
	}

	return len(m.Scores) > len(member.Scores)
}
