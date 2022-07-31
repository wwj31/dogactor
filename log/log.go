package log

import (
	"github.com/wwj31/dogactor/l"
)

var SysLog = l.New(l.Option{
	Level:          l.WarnLevel,
	LogPath:        "./syslog",
	FileName:       "sys.err.log",
	FileMaxAge:     15,
	FileMaxSize:    100,
	FileMaxBackups: 50,
	DisplayConsole: true,
	Skip:           1,
})

func init() {
	SysLog.Color(l.Gray)
}
