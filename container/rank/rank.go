package rank

import (
	"bytes"
	"errors"
	"fmt"
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
	allLen :=int64(len(s.members))
	buffer.Write((*(*[8]byte)(unsafe.Pointer(&allLen)))[:])

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
		s.members[string(key)] = member
		s.skiplist.Insert(member)
	}

	return nil
}