package log

type Logger struct {
	logLevel int32
	_default map[string]interface{}
}

func New(lv int32) *Logger {
	return &Logger{logLevel: lv}
}

func NewWithDefault(lv int32, _default map[string]interface{}) *Logger {
	logger := &Logger{logLevel: lv}
	if len(_default) > 0 {
		logger._default = _default
	}
	return logger
}

func NewWithDefaultAndLogger(l *Logger, _default map[string]interface{}) *Logger {
	logger := NewWithDefault(l.logLevel, _default)
	if len(l._default) > 0 {
		if len(logger._default) == 0 {
			logger._default = make(map[string]interface{})
		}
		for k, v := range l._default {
			logger._default[k] = v
		}
	}
	return logger
}

func (s *Logger) buildDefault(builder *Builder) *Builder {
	if len(s._default) == 0 {
		return builder
	}

	for k, v := range s._default {
		if builder == nil {
			builder = KV(k, v)
		} else {
			builder.KV(k, v)
		}
	}
	return builder
}

func (s *Logger) KV(key string, value interface{}) *Builder {
	builder := KV(key, value)
	builder.level = s.logLevel
	s.buildDefault(builder)
	return builder
}

func (s *Logger) KVs(fields Fields) *Builder {
	builder := KVs(fields)
	builder.level = s.logLevel
	s.buildDefault(builder)
	return builder
}

func (s *Logger) Debug(msg string) {
	if Levels[TAG_DEBUG] >= s.logLevel {
		write(s.buildDefault(nil), 0, TAG_DEBUG, msg)
	}
}
func (s *Logger) Info(msg string) {
	if Levels[TAG_INFO] >= s.logLevel {
		write(s.buildDefault(nil), 0, TAG_INFO, msg)
	}
}
func (s *Logger) Warn(msg string) {
	if Levels[TAG_WARN] >= s.logLevel {
		write(s.buildDefault(nil), 0, TAG_WARN, msg)
	}
}
func (s *Logger) Error(msg string) {
	if Levels[TAG_ERROR] >= s.logLevel {
		write(s.buildDefault(nil), 0, TAG_ERROR, msg)
	}
}
func (s *Logger) Fatal(msg string) {
	if Levels[TAG_FATAL] >= s.logLevel {
		write(s.buildDefault(nil), 0, TAG_FATAL, msg)
	}
}
func (s *Logger) DebugStack(stack int, msg string) {
	if Levels[TAG_DEBUG] >= s.logLevel {
		write(s.buildDefault(nil), stack, TAG_DEBUG, msg)
	}
}
func (s *Logger) InfoStack(stack int, msg string) {
	if Levels[TAG_INFO] >= s.logLevel {
		write(s.buildDefault(nil), stack, TAG_INFO, msg)
	}
}
func (s *Logger) WarnStack(stack int, msg string) {
	if Levels[TAG_WARN] >= s.logLevel {
		write(s.buildDefault(nil), stack, TAG_WARN, msg)
	}
}
func (s *Logger) ErrorStack(stack int, msg string) {
	if Levels[TAG_ERROR] >= s.logLevel {
		write(s.buildDefault(nil), stack, TAG_ERROR, msg)
	}
}
func (s *Logger) FatalStack(stack int, msg string) {
	if Levels[TAG_FATAL] >= s.logLevel {
		write(s.buildDefault(nil), stack, TAG_FATAL, msg)
	}
}
