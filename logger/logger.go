package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
	"reflect"
	"time"
	"unsafe"
)

type Option struct {
	Level          Level       // 日志级别
	LogPath        string      // 日志保存路径
	FileName       string      // 日志文件名称
	FileMaxAge     int         // 文件保存时间(天)
	FileMaxSize    int         // 文件切割大小MB
	FileMaxBackups int         // 最大备份数量
	DisplayConsole bool        // 是否在控制台显示
	Skip           int         // 跳过的栈底
	ExtraWriter    []io.Writer // 扩展输出
}

func New(opt Option) *Logger {
	var (
		lj      *lumberjack.Logger
		writers []io.Writer
	)

	if opt.LogPath != "" {
		//lumberjack
		lj = &lumberjack.Logger{
			Filename:   path.Join(opt.LogPath, opt.FileName),
			MaxSize:    opt.FileMaxSize,
			MaxBackups: opt.FileMaxBackups,
			MaxAge:     opt.FileMaxAge, //days
			Compress:   true,           // disabled by default
		}
		writers = append(writers, lj)
	}

	if opt.DisplayConsole {
		writers = append(writers, os.Stdout)
	}

	writers = append(writers, opt.ExtraWriter...)
	//output = bufio.NewWriter(io.MultiWriter(writers...))

	// zap
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncodeTime = func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		defaultFormat := "2006-01-02 15:04:05.000"
		defaultFormat = time.RFC3339
		encoder.AppendString(t.Format(defaultFormat))
	}

	encoder := zapcore.NewConsoleEncoder(cfg)
	//encoder := zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig())
	//core := zapcore.NewCore(encoder, &sync{Writer: output}, opt.Level)
	zcore := zapcore.NewCore(encoder, zapcore.AddSync(io.MultiWriter(writers...)), opt.Level)
	p := (*ioCore)(unsafe.Pointer(reflect.ValueOf(zcore).Pointer()))

	sugar := zap.New(zcore,
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCallerSkip(opt.Skip),
		zap.AddCaller(),
	).Sugar()

	logger := &Logger{
		Option:  opt,
		rotater: lj,
		sugar:   sugar,
		core:    p,
	}

	loggers = append(loggers, logger)
	return logger
}

type Logger struct {
	Option
	rotater *lumberjack.Logger
	sugar   *zap.SugaredLogger
	core    *ioCore
	color   TColor
	defMsg  string
}

type ioCore struct {
	zapcore.LevelEnabler
	enc zapcore.Encoder
	out zapcore.WriteSyncer
}

func (s *Logger) Close() {
	if s == nil {
		return
	}
	_ = s.sugar.Sync()
	_ = s.rotater.Close()
}

func (s *Logger) DefaultMsg(msg string) *Logger {
	if s == nil {
		return nil
	}
	s.defMsg = msg + " "
	return s
}

func (s *Logger) Storage(lv Level) *Logger {
	if s == nil {
		return nil
	}
	s.core.LevelEnabler = lv
	return s
}

func (s *Logger) Level(lv Level) *Logger {
	if s == nil {
		return nil
	}
	s.core.LevelEnabler = lv
	return s
}

func (s *Logger) Color(c TColor) *Logger {
	if s == nil {
		return nil
	}
	s.color = c
	return s
}

func (s *Logger) CleanColor() *Logger {
	if s == nil {
		return nil
	}
	s.color = 0
	return s
}

func (s *Logger) Debugf(msg string, args ...interface{}) {
	if s == nil {
		return
	}
	msg = color[s.color].B + s.defMsg + msg + color[s.color].E
	s.sugar.Debugf(msg, args...)
}

func (s *Logger) Infof(msg string, args ...interface{}) {
	if s == nil {
		return
	}
	msg = color[s.color].B + s.defMsg + msg + color[s.color].E
	s.sugar.Infof(msg, args...)
}

func (s *Logger) Warnf(msg string, args ...interface{}) {
	if s == nil {
		return
	}
	msg = color[s.color].B + s.defMsg + msg + color[s.color].E
	s.sugar.Warnf(msg, args...)
}

func (s *Logger) Errorf(msg string, args ...interface{}) {
	if s == nil {
		return
	}
	msg = color[s.color].B + s.defMsg + msg + color[s.color].E
	s.sugar.Errorf(msg, args...)
}

func (s *Logger) Debugw(msg string, args ...interface{}) {
	if s == nil {
		return
	}
	msg = color[s.color].B + s.defMsg + msg + color[s.color].E
	s.sugar.Debugw(msg, args...)
}

func (s *Logger) Infow(msg string, args ...interface{}) {
	if s == nil {
		return
	}
	msg = color[s.color].B + s.defMsg + msg + color[s.color].E
	s.sugar.Infow(msg, args...)
}

func (s *Logger) Warnw(msg string, args ...interface{}) {
	if s == nil {
		return
	}
	msg = color[s.color].B + s.defMsg + msg + color[s.color].E
	s.sugar.Warnw(msg, args...)
}

func (s *Logger) Errorw(msg string, args ...interface{}) {
	if s == nil {
		return
	}
	msg = color[s.color].B + s.defMsg + msg + color[s.color].E
	s.sugar.Errorw(msg, args...)
}
