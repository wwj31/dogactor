package rank

import (
	"fmt"
	"testing"
)

func TestRank(t *testing.T) {
	rank := New()
	rank.Add("张三", Score(43, 1))
	rank.Add("李四", Score(12))
	rank.Add("王五", Score(678))
	rank.Add("赵六", Score(1, 3))
	rank.Add("孙七", Score(1, 2))
	rank.Add("周八", Score(12))
	rank.Add("吴九", Score(43))
	rank.Add("郑十", Score(67))
	rank.Add("郑十", Score(68))
	fmt.Println(rank.Get(1, 100))

	rank.Add("孙七", Score(1, 5))
	fmt.Println(rank.Get(1, 100))

	mem1 := rank.Get(5)
	fmt.Println(mem1[0].Key)
}
