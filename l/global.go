package l

var gLogger *Logger
var loggers []*Logger

func Init(option Option) {
	gLogger = New(option)
	loggers = append(loggers, gLogger)
}

func Close() {
	for _, logger := range loggers {
		logger.Close()
	}
}
func Color(c TColor) *Logger {
	gLogger.color = c
	return gLogger
}

func CleanColor() *Logger {
	gLogger.color = 0
	return gLogger
}

func Debugf(msg string, args ...interface{}) {
	msg = color[gLogger.color].B + msg + color[gLogger.color].E
	gLogger.sugar.Debugf(msg, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	msg = color[gLogger.color].B + msg + color[gLogger.color].E
	gLogger.sugar.Debugw(msg, keysAndValues...)
}

func Infof(msg string, args ...interface{}) {
	msg = color[gLogger.color].B + msg + color[gLogger.color].E
	gLogger.sugar.Infof(msg, args...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	msg = color[gLogger.color].B + msg + color[gLogger.color].E
	gLogger.sugar.Infow(msg, keysAndValues...)
}

func Warnf(msg string, args ...interface{}) {
	msg = color[gLogger.color].B + msg + color[gLogger.color].E
	gLogger.sugar.Warnf(msg, args...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	msg = color[gLogger.color].B + msg + color[gLogger.color].E
	gLogger.sugar.Warnw(msg, keysAndValues...)
}

func Errorf(msg string, args ...interface{}) {
	msg = color[gLogger.color].B + msg + color[gLogger.color].E
	gLogger.sugar.Errorf(msg, args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	msg = color[gLogger.color].B + msg + color[gLogger.color].E
	gLogger.sugar.Errorw(msg, keysAndValues...)
}
