package log

import (
	"github.com/wwj31/dogactor/logger"
)

var SysLog *logger.Logger

func init() {
	SysLog = logger.New(logger.Option{
		Level:          logger.InfoLevel,
		LogPath:        "./syslog",
		FileName:       "sys.err.log",
		FileMaxAge:     15,
		FileMaxSize:    100,
		FileMaxBackups: 50,
		DisplayConsole: true,
		Skip:           1,
	})

	SysLog.Color(logger.Gray)
}
