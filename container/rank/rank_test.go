package rank

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// 重复key替换测试
func TestAdd(t *testing.T) {
	rank := New("", time.Now())
	rank.Add("a", 43)
	rank.Add("b", 12)
	rank.Add("c", 678)
	fmt.Println("all:", rank.Get())
	rank.Add("d", 54, 1000)
	rank.Add("dd", 54)
	rank.Add("e", 54, 2000)
	//rank.Add("f", 67, 1111)
	//rank.Add("g", 67, 2222)
	//rank.Add("h", 67, 1000)

	rank.Add("g", 68, 1000)
	rank.Add("h", 1, 5)
	fmt.Println("all:", rank.Get())

	fmt.Println(rank.GetByMemberID("g"))
	fmt.Println(rank.GetByMemberID("c"))
}

// 积分区间测试
func TestGetByScore(t *testing.T) {
	rank := New("", time.Now())
	for i := 0; i < 100; i++ {
		rank.Add(fmt.Sprintf("number:%v ", i), rand.Int63n(1000), rand.Int63n(1000))
	}
	rank.Add(fmt.Sprintf("number:%v ", 100), 100, 100)
	rank.Add(fmt.Sprintf("number:%v ", 102), 100, 100)
	rank.Add(fmt.Sprintf("number:%v ", 500), 500, 500)
	r := rank.Get()
	for _, v := range r {
		fmt.Println(v)
	}

	fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<100 ~ 500>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	section := rank.GetByScore(100, 500)
	for _, v := range section {
		fmt.Println(v)
	}
}

func TestRankFind(t *testing.T) {
	rank := New("", time.Now())
	rank.Add("a", 100)
	rank.Add("b", 101)
	rank.Add("c", 90)
	rank.Add("d", 94)
	rank.Add("e", 94)
	fmt.Println("all:", rank.Get())
	rank.Del("e")
	fmt.Println("del e:", rank.Get())
	//fmt.Println("find 100:", rank.skipList.Find(Member{Scores: []Score{100}}))
	//fmt.Println("find 100:", rank.skipList.Find(Member{Scores: []Score{101}}))
	//fmt.Println("all:", rank.Get())
}
