package rank

import (
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/wwj31/dogactor/container/skiplist"
	"github.com/wwj31/dogactor/tools"
)

type (
	Rank struct {
		Key      string
		SkipList *skiplist.SkipList `json:"-"`
		Members  map[string]Member
		ExpireAt time.Time // 零值表示不过期
		ScoreInc int64
		Modify   bool `json:"-"`
	}
)

func New(key string, expireAt time.Time) *Rank {
	return &Rank{
		Key:      key,
		SkipList: skiplist.New(),
		Members:  make(map[string]Member, 10),
		ExpireAt: expireAt,
	}
}

func (r *Rank) Len() int {
	return len(r.Members)
}

// Add 多分排行，按照scores传入顺序，优先级由高到低排序
func (r *Rank) Add(memberID string, scores ...Score) (*Rank, error) {
	if len(scores) == 0 {
		return nil, fmt.Errorf("add failed scores len == 0")
	}

	if member, ok := r.Members[memberID]; ok {
		// 分数没有变化，无需更改排名
		if reflect.DeepEqual(scores, r.ScoreWithoutInc(member.Scores)) {
			return r, nil
		}

		// 分数有变，先删老数据再加新数据
		r.SkipList.Delete(member)
	}

	r.Members[memberID] = r.SkipList.Insert(Member{
		ID:     memberID,
		Scores: r.Score(scores...),
	}).Value.(Member)

	r.Modify = true
	return r, nil
}

// Get rankSection 名次区间
// Rank.Get() 获得全部名次
// Rank.Get(1) 获得1名
// Rank.Get(1,100) 获得1～100名
func (r *Rank) Get(rankSection ...int64) []Member {
	var (
		top     int64
		bottom  int64
		members = make([]Member, 0)
	)

	if len(r.Members) == 0 {
		return members
	}
	if len(rankSection) > 0 {
		top = rankSection[0]
	}
	if len(rankSection) > 1 {
		bottom = rankSection[1]
	}
	if top == 0 {
		top = 1
		bottom = math.MaxInt64
	}

	ele := r.SkipList.GetElementByRank(int(top))
	if ele == nil || ele.Value == nil {
		return members
	}
	members = append(members, ele.Value.(Member))
	top++
	for ; top <= bottom; top++ {
		val := ele.Next()
		if val == nil {
			break
		}
		members = append(members, val.Value.(Member))
		ele = val
	}
	return members
}

// GetByMemberID 查找key的名次、分数，找不到返回0
func (r *Rank) GetByMemberID(id string) (order int, member Member) {
	member, ok := r.Members[id]
	if !ok {
		return
	}
	return r.SkipList.GetRank(member), member
}

// GetByScore 获得区间所有排名信息，闭区间，只能获得权重最大的积分区间
func (r *Rank) GetByScore(floorScore, roofScore Score) []Member {
	members := make([]Member, 0)
	roof := Member{Scores: []Score{roofScore, 0, 0, 0, 0, 0}}
	floor := Member{Scores: []Score{floorScore, math.MaxInt64}}
	for rf := r.SkipList.Find(roof); rf != nil; {
		if rf.Value.Less(floor) {
			members = append(members, rf.Value.(Member))
		}
		rf = rf.Next()
	}
	return members
}

func (r *Rank) Del(memberID string) *Rank {
	member, ok := r.Members[memberID]
	if !ok {
		return r
	}

	delete(r.Members, memberID)
	r.SkipList.Delete(member)
	r.Modify = true
	return r
}

func (r *Rank) Expire() bool {
	if r.ExpireAt.IsZero() {
		return false
	}

	return tools.Now().After(r.ExpireAt)
}

// Score 传入分数依次作为排名权重,所有权重相同根据插入顺序先到的排前
func (r *Rank) Score(scores ...Score) (result []Score) {
	for _, i64 := range scores {
		result = append(result, i64)
	}

	r.ScoreInc = tools.Max(0, r.ScoreInc+1)
	result = append(result, r.ScoreInc)
	return result
}

func (r *Rank) ScoreWithoutInc(scores []Score) (result []Score) {
	// 排行榜分数，至少2个参数
	if len(scores) <= 1 {
		return
	}

	for i := 0; i < len(scores)-1; i++ {
		result = append(result, scores[i])
	}
	return
}
