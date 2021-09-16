package rank

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"reflect"
	"testing"
)

// 重复key替换测试
func TestAdd(t *testing.T) {
	rank := New()
	rank.Add("张三", Score(43))
	rank.Add("李四", Score(12))
	rank.Add("王五", Score(678))
	rank.Add("赵六", Score(54, 1000))
	rank.Add("孙七", Score(54, 2000))
	rank.Add("周八", Score(67, 1111))
	rank.Add("吴九", Score(67, 2222))
	rank.Add("郑十", Score(67, 1000))

	rank.Add("郑十", Score(68, 1000))
	rank.Add("孙七", Score(1, 5))
	fmt.Println("全排名:", rank.Get())

	fmt.Println(rank.GetByKey("周八"))
	fmt.Println(rank.GetByKey("张三"))
}

// 积分区间测试
func TestGetByScore(t *testing.T) {
	rank := New()
	for i := 0; i < 100; i++ {
		rank.Add(fmt.Sprintf("number:%v ", i), Score(rand.Int63n(1000), rand.Int63n(1000)))
	}
	r := rank.Get()
	for _, v := range r {
		fmt.Println(v)
	}

	fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<100 ~ 500>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	section := rank.GetByScore([]int64{100}, []int64{200})
	for _, v := range section {
		fmt.Println(v)
	}
}

// 序列化测试
func TestMarshal(t *testing.T) {
	rank := New()
	for i := 0; i < 1000000; i++ {
		rank.Add(fmt.Sprintf(uuid.New().String()), Score(rand.Int63n(10000), rand.Int63n(10000), rand.Int63n(10000)))
	}
	data := rank.Marshal()
	fmt.Println(fmt.Sprintf("bytes size:%.3f MB", float64(len(data))/(1024*1024)))

	rank2 := New()
	_ = rank2.UnMarshal(data)
	v := rank.skiplist.Front()
	v2 := rank2.skiplist.Front()
	for v != nil {
		if v.Value.(Member).Key != v2.Value.(Member).Key || !reflect.DeepEqual(v.Value.(Member).Scores, v2.Value.(Member).Scores) {
			panic("error")
		}

		v = v.Next()
		v2 = v2.Next()
	}
	if v2 != nil {
		panic("error")
	}
	fmt.Println("it's ok!")
}
