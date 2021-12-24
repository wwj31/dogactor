package test

import (
	"github.com/wwj31/dogactor/l"
	"github.com/wwj31/dogactor/tools"
	"testing"
)

func TestGLogger(t *testing.T) {
	l.Init(l.Option{
		Level:          l.DebugLevel,
		LogPath:        "./testDir",
		FileName:       "test.log",
		FileMaxAge:     1,
		FileMaxSize:    1,
		FileMaxBackups: 3,
		DisplayConsole: true,
	})

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
	l.Color(l.Red).Debugw("this is debug",
		"k", "v",
		"val", 123,
		"uuid", tools.UUID())

	l.Infow("this is info",
		"k", "v",
		"val", 123,
		"uuid", tools.UUID())

	l.Warnw("this is warn",
		"k", "v",
		"val", 123,
		"uuid", tools.UUID())

	ln.DefaultMsg("default msg").Errorw("this is error",
		"k", "v",
		"val", 123,
		"uuid", tools.UUID())

	ln.Color(l.Gray).Errorf("this is error %v %v",
		"uuid", tools.UUID())
}
