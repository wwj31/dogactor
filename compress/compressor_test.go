package compress

import (
	"fmt"
	"testing"
)


//func TestCompressor(t *testing.T) {
//	bytes,_ := Compress2([]byte(str))
//	data,_ := Uncompress(bytes)
//	str2 := string(data)
//	fmt.Println(str2)
//}
var teststr string
func init()  {
	teststr = str+str+str+str+str+str+str+str+str+str+str

	// teststr 1MB
	fmt.Println(len(teststr))
}

func BenchmarkCompresspg(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compresspg([]byte(teststr),100*1024,2)
	}
}

func BenchmarkCompress2(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compress([]byte(teststr))
	}
}