package log

import (
	"github.com/wwj31/dogactor/l"
)

var SysLog = l.New(l.Option{
	Level:          l.InfoLevel,
	LogPath:        "./syslog",
	FileName:       "sys.log",
	FileMaxAge:     15,
	FileMaxSize:    512,
	FileMaxBackups: 10,
	DisplayConsole: true,
})
