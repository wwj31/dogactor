package rank

import (
	"bytes"
	"github.com/wwj31/dogactor/container/skiplist"
	"reflect"
	"unsafe"
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
	m, ok := s.members[key]

	// Fast path
	if reflect.DeepEqual(m.Scores,scores){
		return
	}

	// Slow path
	if ok{
		s.skiplist.Delete(m)
	}
	m.Scores = scores
	m.Key = key
	s.members[key] = m
	s.skiplist.Insert(m)
	return
}


// need top >= 1
func (s *Rank) Get(top int, bottom ...int) []Member {
	members := make([]Member, 0)
	if len(s.members) == 0 {
		return members
	}
	ele := s.skiplist.GetElementByRank(top)
	if ele == nil || ele.Value == nil {
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

func (s *Rank) Marshal() []byte{
	buffer := bytes.NewBuffer([]byte{})
	l:=int32(len(s.members))
	buffer.Write((*(*[4]byte)(unsafe.Pointer(&l)))[:])

	for key,member := range s.members{
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

func (s *Rank) UnMarshal(data []byte) error{
	buffer := bytes.NewBuffer(data)
	l := [4]byte{}
	buffer.Read(l[:])

	ml := *(*int32)(unsafe.Pointer(&l))
	for i:=int32(0);i<ml;i++{
		k := [4]byte{}
		n := [4]byte{}
		buffer.Read(k[:])
		buffer.Read(n[:])
		keylen := *(*int32)(unsafe.Pointer(&k))
		nlen := *(*int32)(unsafe.Pointer(&n))

		key := make([]byte,keylen)
		buffer.Read(key)
		member := Member{
			Key: string(key),
			Scores:make([]num,0),
		}
		for i:=int32(0);i < nlen;i++{
			number := [8]byte{}
			buffer.Read(number[:])
			member.Scores = append(member.Scores,*(*num)(unsafe.Pointer(&number)))
		}
		s.members[string(key)] = member
		s.skiplist.Insert(member)
	}

	return nil
}