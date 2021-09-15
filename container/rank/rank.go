package rank

import (
	"github.com/wwj31/dogactor/container/skiplist"
)

type (
	Rank struct {
		skiplist *skiplist.SkipList
		members  map[string]Member
	}

	// 分数类型
	num = int64
)

func New() *Rank {
	return &Rank{
		skiplist: skiplist.New(),
		members:  make(map[string]Member, 10),
	}
}

func (s *Rank)Len() int {
	return s.skiplist.Len()
}

func (s *Rank) Add(key string, scores []num) {
	oldm, ok := s.members[key]

	// Fast path
	if len(oldm.Scores) == len(scores) {
		c := true
		for i := 0; i < len(oldm.Scores); i++ {
			if oldm.Scores[i] != scores[i] {
				c = false
				break
			}
		}
		if c {
			return
		}
	}

	// Slow path
	newm := Member{
		Key:    key,
		Scores: scores,
	}
	s.members[key] = newm
	if !ok {
		s.skiplist.Insert(newm)
		return
	}

	s.skiplist.Delete(oldm)
	s.skiplist.Insert(newm)
	return
}


// need top >= 1
func (s *Rank) Get(top int, bottom ...int) []Member {
	members := make([]Member, 0)
	if len(s.members) == 0 {
		return members
	}
	ele := s.skiplist.GetElementByRank(top)
	if ele.Value == nil {
		return members
	}

	members = append(members, ele.Value.(Member))
	if len(bottom) == 0 || bottom[0] <= top {
		return members
	}
	top++

	for ; top <= bottom[0]; top++ {
		val := ele.Next()
		if val == nil {
			break
		}
		members = append(members, val.Value.(Member))
		ele = val
	}
	return members
}

func (s *Rank) Del(key string) {
	member, ok := s.members[key]
	if !ok {
		return
	}
	delete(s.members, key)
	s.skiplist.Delete(member)
}