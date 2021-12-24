package l

import "time"

var loggers []*Logger

func init() {
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for range ticker.C {
			for _, l := range loggers {
				_ = l.sugar.Sync()
			}
		}
	}()
}

func Close() {
	for _, logger := range loggers {
		logger.Close()
	}
}
