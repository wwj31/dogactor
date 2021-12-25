package test

import (
	"github.com/wwj31/dogactor/l"
	"github.com/wwj31/dogactor/tools"
	"testing"
)

func TestGLogger(t *testing.T) {
	GOGO()
	l.Close()
}

func GOGO() {
	ln := l.New(l.Option{
		Level:          l.DebugLevel,
		LogPath:        "./testLog",
		FileName:       "test.log",
		FileMaxAge:     1,
		FileMaxSize:    1,
		FileMaxBackups: 3,
		DisplayConsole: true,
	})

	ln.DefaultMsg("default msg").Errorw("this is error",
		"k", "v",
		"val", 123,
		"uuid", tools.UUID())

	ln.Color(l.Gray).Errorf("this is error %v %v",
		"uuid", tools.UUID())
}
