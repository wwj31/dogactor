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
		FileMaxSize:    10,
		FileMaxBackups: 30,
		DisplayConsole: true,
	})

	for {
		ln.Errorw("this is error",
			"k", "v",
			"val", 123,
			"uuid", tools.UUID())
	}
}
