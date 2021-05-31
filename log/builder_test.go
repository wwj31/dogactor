package log

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

/*
BenchmarkSprintfWrite
BenchmarkSprintfWrite-8      	  668715	      1930 ns/op
BenchmarkStringsBuilder1
BenchmarkStringsBuilder1-8   	 1000000	      1117 ns/op
BenchmarkStringsBuilder2
BenchmarkStringsBuilder2-8   	 1000000	      1159 ns/op
*/

func SprintfWrite(t time.Time, color, tag string, file string, line int, msg string) {
	info := fmt.Sprintf("%s %s[%s] %s:%d %s\n", t.Format("2006-01-02 15:04:05"), color, tag, file, line, msg)
	datas <- info
}

var datas = make(chan string, 10000000)

func StringsBuilderWrite1(t time.Time, color, tag string, file string, line int, msg string) {
	sb := &Builder{buf: make([]byte, 0, 128)}
	sb.WriteString(t.Format("2006-01-02 15:04:05"))
	sb.WriteString(color)
	sb.WriteString(" [")
	sb.WriteString(tag)
	sb.WriteString("] ")
	sb.WriteString(file)
	sb.WriteString(":")
	sb.WriteString(strconv.Itoa(line))
	sb.WriteString(" ")
	sb.WriteString(msg)
	sb.WriteString("\n")
	sb.Bytes()
}

func StringsBuilderWrite2(t time.Time, color, tag string, file string, line int, msg string) {
	sb := &strings.Builder{}
	sb.WriteString(t.Format("2006-01-02 15:04:05"))
	sb.WriteString(color)
	sb.WriteString(" [")
	sb.WriteString(tag)
	sb.WriteString("] ")
	sb.WriteString(file)
	sb.WriteString(":")
	sb.WriteString(strconv.Itoa(line))
	sb.WriteString(" ")
	sb.WriteString(msg)
	sb.WriteString("\n")
	_ = []byte(sb.String())
}

func BenchmarkSprintfWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SprintfWrite(time.Now(), "", "D", "aaaaaaa", 10, "client msg sessionId=747010017 msg=GET_TILEINFO cost=77.107µsclient msg sessionId=747010017 msg=GET_TILEINFO cost=77.107µs")
	}
}

func BenchmarkStringsBuilder1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StringsBuilderWrite1(time.Now(), "", "D", "aaaaaaa", 10, "client msg sessionId=747010017 msg=GET_TILEINFO cost=77.107µsclient msg sessionId=747010017 msg=GET_TILEINFO cost=77.107µs")
	}
}

func BenchmarkStringsBuilder2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StringsBuilderWrite2(time.Now(), "", "D", "aaaaaaa", 10, "client msg sessionId=747010017 msg=GET_TILEINFO cost=77.107µsclient msg sessionId=747010017 msg=GET_TILEINFO cost=77.107µs")
	}
}
