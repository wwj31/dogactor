package log

import (
	"flag"
	"github.com/wwj31/dogactor/logger"
	"go.uber.org/zap/zapcore"
)

var (
	sysLogPath = flag.String("syslogpath", "./syslog", "system log path")
	sysLogLv   = flag.Int("sysloglevel", int(logger.InfoLevel), "system log level")
)

var SysLog *logger.Logger

func init() {
	flag.Parse()

	SysLog = logger.New(logger.Option{
		Level:          zapcore.Level(*sysLogLv),
		LogPath:        *sysLogPath,
		FileName:       "sys.err.log",
		FileMaxAge:     15,
		FileMaxSize:    100,
		FileMaxBackups: 50,
		DisplayConsole: true,
		Skip:           1,
	})

	SysLog.Color(logger.Gray)
}
