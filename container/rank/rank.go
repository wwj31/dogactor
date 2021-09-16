package rank

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"reflect"
	"time"
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

func (r *Rank)Len() int {
	return r.skiplist.Len()
}


var _inc int64

// Score 作为 Rank.Add 第二参数，传入分数依次作为排名权重
func Score(scores ...int64) []num {
	nums := []num{}
	for _, i64 := range scores{
		nums = append(nums, num(i64))
	}

	nums = append(nums, num(math.MaxInt64-(int64(time.Now().Nanosecond())+_inc)))
	_inc++
	return nums
}

// Add 例子：
// Rank.Add("xxxx",rank.Score(999)) 单分排行
// Rank.Add("xxxx",rank.Score(999,123,456)) 多分排行
func (r *Rank) Add(key string, scores []num) *Rank{
	m, ok := r.members[key]

	// Fast path
	if reflect.DeepEqual(m.Scores,scores){
		return r
	}

	// Slow path
	if ok{
		r.skiplist.Delete(m)
	}
	m.Scores = scores
	m.Key = key
	r.members[key] = m
	r.skiplist.Insert(m)
	return r
}

// Get rankSection 名次区间
// 例子：
// members := Rank.Get() 获得全部名次
// members := Rank.Get(1) 获得1名
// members := Rank.Get(3) 获得3名
// members := Rank.Get(1,100) 获得1～100名
func (r *Rank) Get(rankSection ...int) []Member {
	var (
		top int
		bottom int
		members = make([]Member, 0)
	)

	if len(r.members) == 0 {
		return members
	}
	if len(rankSection) > 0{
		top = rankSection[0]
	}
	if len(rankSection) > 1{
		bottom = rankSection[1]
	}
	if top == 0{
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

// GetByScore scoreSection 分数区间
// 例子：
// Rank.GetByScore() 获得全部名次
// Rank.GetByScore(999) 获得分数为999的集合
// Rank.GetByScore(100~999) 获得分数为100~999区间的集合
func (r *Rank) GetByScore(floorScores, roofScores []int64) []Member{
	members := make([]Member,0)
	if roofScores == nil || floorScores == nil {
		return members
	}

	for k,v := range roofScores{
		roofScores[k] = v+1
	}
	floor := Member{Scores: *((*[]num)(unsafe.Pointer(&floorScores)))}
	roof := Member{Scores: *((*[]num)(unsafe.Pointer(&roofScores)))}


	for rf := r.skiplist.Find(roof);rf != nil;rf = rf.Next(){
		if rf.Value.Less(floor){
			members = append(members,rf.Value.(Member))
		}
	}
	return members
}

func (r *Rank) Del(key string) *Rank{
	member, ok := r.members[key]
	if !ok {
		return r
	}
	delete(r.members, key)
	r.skiplist.Delete(member)
	return r
}

func (r *Rank) Marshal() []byte{
	buffer := bytes.NewBuffer([]byte{})
	allLen :=int64(len(r.members))
	buffer.Write((*(*[8]byte)(unsafe.Pointer(&allLen)))[:])

	for key,member := range r.members{
		keylen := int32(len(key))
		nlen := int32(len(member.Scores))
		buffer.Write((*(*[4]byte)(unsafe.Pointer(&keylen)))[:])
		buffer.Write((*(*[4]byte)(unsafe.Pointer(&nlen)))[:])

		buffer.WriteString(key)
		for _,n := range member.Scores{
			buffer.Write(((*[8]byte)(unsafe.Pointer(&n)))[:])
		}
	}
	return buffer.Bytes()
}

func (r *Rank) UnMarshal(data []byte) error{
	buffer := bytes.NewBuffer(data)

	allLenByte8 := [8]byte{}
	if n,err := buffer.Read(allLenByte8[:]);n != 8 || err != nil{
		return errors.New(fmt.Sprintf("read allLenByte8 err n:%v err:%v",n,err))
	}
	allLen := *(*int64)(unsafe.Pointer(&allLenByte8))

	for i:=int64(0);i< allLen;i++{
		keyLenBytes4 := [4]byte{}
		if n,err := buffer.Read(keyLenBytes4[:]);n != 4 || err != nil{
			return errors.New(fmt.Sprintf("read keyLenBytes4 err n:%v err:%v",n,err))
		}
		nLenBytes4 := [4]byte{}
		if n,err := buffer.Read(nLenBytes4[:]);n != 4 || err != nil{
			return errors.New(fmt.Sprintf("read nLenBytes4 err n:%v err:%v",n,err))
		}
		keylen := *(*int32)(unsafe.Pointer(&keyLenBytes4))
		nlen := *(*int32)(unsafe.Pointer(&nLenBytes4))

		key := make([]byte,keylen)
		if n,err := buffer.Read(key);n != int(keylen) || err != nil{
			return errors.New(fmt.Sprintf("read key err n:%v int(keylen):%v err:%v",n,int(keylen),err))
		}
		member := Member{
			Key: string(key),
			Scores:make([]num,0,nlen),
		}
		for i:=int32(0);i < nlen;i++{
			number := [8]byte{}
			if n,err := buffer.Read(number[:]);n != 8 || err != nil{
				return errors.New(fmt.Sprintf("read number err n:%v err:%v",n,err))
			}
			member.Scores = append(member.Scores,*(*num)(unsafe.Pointer(&number)))
		}
		r.members[string(key)] = member
		r.skiplist.Insert(member)
	}

	return nil
}