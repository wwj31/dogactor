package log

import (
	"testing"
)

func BenchmarkBuilder_Log(b *testing.B) {
	test1()
}

func test1() {
	test2()
}
func test2() {
	test3()
}
func test3() {
	test4()
}
func test4() {
	KV("aaa", 123).KVs(Fields{"c1": "c1", "123": 4444}).DebugStack(0, "测试日志")
}
