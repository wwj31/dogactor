package network

import (
	"fmt"
	"testing"
)

func TestNewBaseNetHandler(t *testing.T) {
	handler := &BaseNetHandler{}
	fmt.Println(handler.session() == nil)
}
