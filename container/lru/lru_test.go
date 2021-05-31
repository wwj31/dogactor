package lru

import (
	"fmt"
	"testing"
)

func TestLruContainer_PushFront(t *testing.T) {
	lc := NewLRU(10)
	for idx := 0; idx < 100; idx++ {
		lc.PushFront(idx%10, idx)
	}
	fmt.Println(lc.Len())

	fmt.Println(lc.Get(5))
	fmt.Println(lc.Get(4))

	lc.Del(7)

	for idx := 0; idx < 10; idx++ {
		fmt.Println(lc.PopTail())
	}

	fmt.Println(lc.Get(4))
	fmt.Println(lc.Get(5))

	fmt.Println(lc.Len())
}

var glc = NewLRU(1000)
var idx int

func BenchmarkLruContainer_PushFront(b *testing.B) {
	for i := 0; i < b.N; i++ {
		glc.PushFront(idx, idx)
	}
}

func BenchmarkLruContainer_Get(b *testing.B) {
	for i := 0; i < b.N; i++ {
		glc.Get(idx)
	}
}
