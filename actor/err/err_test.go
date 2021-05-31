package err

import (
	"fmt"
	"testing"
)

func TestErr(t *testing.T) {
	e1 := fmt.Errorf("this is error!")
	e2 := fmt.Errorf("%w param1:%v,param2:%v", e1, 123, "bbb")

	fmt.Print(e2)
	fmt.Println(" ")
}
