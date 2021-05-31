package ringbuffer

import (
	"fmt"
	"log"
	"testing"
)

/*
BenchmarkQueue_Ring
BenchmarkQueue_Ring-8    	30336562	        39.7 ns/op
BenchmarkQueue_Array
BenchmarkQueue_Array-8   	13819460	        86.4 ns/op
*/

var que = NewQueue()
var array = &QueArray{}

func TestQueue_Ring(t *testing.T) {
	a := []int{1, 2, 3}
	log.Println(a[1:3])

	for idx := 0; idx < 32; idx++ {
		que.PushBack(idx)
	}
	for idx := 0; idx < 32; idx++ {
		fmt.Println(que.Front(), que.Back(), que.PopFront())
	}

	for idx := 0; idx < 32; idx++ {
		array.Push(idx)
	}

	for idx := 0; idx < 32; idx++ {
		array.Pop()
		array.Push(idx)
		fmt.Println(cap(array.array))
	}

}

func BenchmarkQueue_Ring(b *testing.B) {
	for idx := 0; idx < b.N; idx++ {
		que.PushBack(idx)
		que.PopFront()
	}
}

func BenchmarkQueue_Array(b *testing.B) {
	for idx := 0; idx < b.N; idx++ {
		array.Push(idx)
		array.Pop()
	}
}
