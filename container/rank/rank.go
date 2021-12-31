package rank

import (
	"bytes"
	"math"
	"reflect"
	"unsafe"

	"github.com/wwj31/dogactor/container/skiplist"
)

type (
	Rank struct {
		skiplist *skiplist.SkipList
		members  map[string]Member
	}
)

func New() *Rank {
	return &Rank{
		skiplist: skiplist.New(),
		members:  make(map[string]Member, 10),
	}
}

func (r *Rank) Len() int {
	return r.skiplist.Len()
}

var _inc int64

// Add Rank.Add("xxxx",999) 单分排行
// Rank.Add("xxxx",999,123,456) 多分排行
func (r *Rank) Add(key string, scores ...int64) *Rank {
	m, ok := r.members[key]

	s := score(scores...)
	// Fast path
	if reflect.DeepEqual(m.Scores, s) {
		return r
	}

	// Slow path
	if ok {
		r.skiplist.Delete(m)
	}
	m.Scores = s
	m.Key = key
	r.members[key] = m
	r.skiplist.Insert(m)
	return r
}

// Get rankSection 名次区间
// Rank.Get() 获得全部名次
// Rank.Get(1) 获得1名
// Rank.Get(1,100) 获得1～100名
func (r *Rank) Get(rankSection ...int) []Member {
	var (
		top     int
		bottom  int
		members = make([]Member, 0)
	)

	if len(r.members) == 0 {
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

	ele := r.skiplist.GetElementByRank(top)
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

// GetByKey 查找key的名次、分数，找不到返回0
func (r *Rank) GetByKey(key string) (int, Member) {
	member, ok := r.members[key]
	if !ok {
		return 0, member
	}
	return r.skiplist.GetRank(member), member
}

// GetByScore scoreSection 分数区间
// Rank.GetByScore([]int64{100},[]int64{900}) 获得分数为100~999区间的集合
func (r *Rank) GetByScore(floorScores, roofScores []int64) []Member {
	members := make([]Member, 0)
	if roofScores == nil || floorScores == nil {
		return members
	}

	// 跳表找的是开区间: (roofScores,∞  所以这里对大值+1
	for k, v := range roofScores {
		roofScores[k] = v + 1
	}

	floor := Member{Scores: *((*[]num)(unsafe.Pointer(&floorScores)))}
	roof := Member{Scores: *((*[]num)(unsafe.Pointer(&roofScores)))}
	for rf := r.skiplist.Find(roof); rf != nil; rf = rf.Next() {
		if rf.Value.Less(floor) {
			members = append(members, rf.Value.(Member))
		} else {
			break
		}
	}
	return members
}

func (r *Rank) Del(key string) *Rank {
	member, ok := r.members[key]
	if !ok {
		return r
	}
	delete(r.members, key)
	r.skiplist.Delete(member)
	return r
}

func (r *Rank) Marshal() []byte {
	buffer := bytes.NewBuffer([]byte{})
	allLen := int64(len(r.members))
	buffer.Write((*(*[8]byte)(unsafe.Pointer(&allLen)))[:])

	for key, member := range r.members {
		keylen := int32(len(key))
		nlen := int32(len(member.Scores))
		buffer.Write((*(*[4]byte)(unsafe.Pointer(&keylen)))[:])
		buffer.Write((*(*[4]byte)(unsafe.Pointer(&nlen)))[:])

		buffer.WriteString(key)
		for _, n := range member.Scores {
			buffer.Write(((*[8]byte)(unsafe.Pointer(&n)))[:])
		}
	}
	return buffer.Bytes()
}

func (r *Rank) UnMarshal(data []byte) error {
	buffer := bytes.NewBuffer(data)

	allLenByte8 := [8]byte{}
	if n, err := buffer.Read(allLenByte8[:]); n != 8 || err != nil {
		return ErrorUnMarshalRead{err: err, n: n}
	}
	allLen := *(*int64)(unsafe.Pointer(&allLenByte8))

	for i := int64(0); i < allLen; i++ {
		keyLenBytes4 := [4]byte{}
		if n, err := buffer.Read(keyLenBytes4[:]); n != 4 || err != nil {
			return ErrorUnMarshalRead{err: err, n: n}
		}
		nLenBytes4 := [4]byte{}
		if n, err := buffer.Read(nLenBytes4[:]); n != 4 || err != nil {
			return ErrorUnMarshalRead{err: err, n: n}
		}
		keylen := *(*int32)(unsafe.Pointer(&keyLenBytes4))
		nlen := *(*int32)(unsafe.Pointer(&nLenBytes4))

		key := make([]byte, keylen)
		if n, err := buffer.Read(key); n != int(keylen) || err != nil {
			return ErrorUnMarshalRead{err: err, n: n}
		}
		member := Member{
			Key:    string(key),
			Scores: make([]num, 0, nlen),
		}
		for i := int32(0); i < nlen; i++ {
			number := [8]byte{}
			if n, err := buffer.Read(number[:]); n != 8 || err != nil {
				return ErrorUnMarshalRead{err: err, n: n}
			}
			member.Scores = append(member.Scores, *(*num)(unsafe.Pointer(&number)))
		}
		r.members[string(key)] = member
		r.skiplist.Insert(member)
	}

	return nil
}
