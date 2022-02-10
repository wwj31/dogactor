package l

var loggers []*Logger

func Close() {
	for _, logger := range loggers {
		logger.Close()
	}
}
