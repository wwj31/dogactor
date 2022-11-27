package test

import (
	"github.com/wwj31/dogactor/logger"
	"github.com/wwj31/dogactor/tools"
	"testing"
	"time"
)

func TestGLogger(t *testing.T) {
	GOGO()
	logger.Close()
}

func GOGO() {
	ln := logger.New(logger.Option{
		Level:          logger.DebugLevel,
		LogPath:        "./testLog",
		FileName:       "test.log",
		FileMaxAge:     1,
		FileMaxSize:    10,
		FileMaxBackups: 30,
		DisplayConsole: true,
	})

	var n int
	for {
		if n == 10 {
			ln.Level(logger.WarnLevel)
		}
		n++
		time.Sleep(time.Second)
		ln.Warnw("this is error",
			"k", "v",
			"val", 123,
			"uuid", tools.XUID())
		ln.Infow("this is info",
			"k", "v",
			"val", 123,
			"uuid", tools.XUID())
		ln.Debugw("this is debug",
			"k", "v",
			"val", 123,
			"uuid", tools.XUID())
	}
}
