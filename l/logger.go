package l

import (
	"bufio"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
	"os"
	"path"

	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Option struct {
	Level          Level  // 日志级别
	LogPath        string // 日志保存路径
	FileName       string // 日志文件名称
	FileMaxAge     int    // 文件保存时间(天)
	FileMaxSize    int    // 文件切割大小MB
	FileMaxBackups int    // 最大备份数量
	DisplayConsole bool   // 是否在控制台显示
}

func New(opt Option) *Logger {
	var (
		output *bufio.Writer
		lj     *lumberjack.Logger
	)
	//lumberjack
	lj = &lumberjack.Logger{
		Filename:   path.Join(opt.LogPath, opt.FileName),
		MaxSize:    opt.FileMaxSize,
		MaxBackups: opt.FileMaxBackups,
		MaxAge:     opt.FileMaxAge, //days
		Compress:   true,           // disabled by default
	}
	writers := []io.Writer{lj}
	if opt.DisplayConsole {
		writers = append(writers, os.Stdout)
	}
	log.SetOutput(io.MultiWriter(writers...))
	output = bufio.NewWriter(log.Writer())

	// zap
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	//encoder := zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, &sync{Writer: output}, opt.Level)
	sugar := zap.New(core,
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCallerSkip(1),
		zap.AddCaller(),
	).Sugar()

	logger := &Logger{
		Option:  opt,
		rotater: lj,
		sugar:   sugar,
	}

	loggers = append(loggers, logger)
	return logger
}

type sync struct {
	*bufio.Writer
}

func (s *sync) Sync() error {
	return s.Writer.Flush()
}

type Logger struct {
	Option
	rotater *lumberjack.Logger
	sugar   *zap.SugaredLogger
	color   TColor
}

func (s *Logger) Close() {
	_ = s.sugar.Sync()
	_ = s.rotater.Close()
}

func (s *Logger) Color(c TColor) *Logger {
	s.color = c
	return s
}

func (s *Logger) CleanColor() *Logger {
	s.color = 0
	return s
}

func (s *Logger) Debugf(msg string, args ...interface{}) {
	msg = color[s.color].B + msg + color[s.color].E
	s.sugar.Debugf(msg, args)
}

func (s *Logger) Infof(msg string, args ...interface{}) {
	msg = color[s.color].B + msg + color[s.color].E
	s.sugar.Infof(msg, args)
}

func (s *Logger) Warnf(msg string, args ...interface{}) {
	msg = color[s.color].B + msg + color[s.color].E
	s.sugar.Warnf(msg, args)
}

func (s *Logger) Errorf(msg string, args ...interface{}) {
	msg = color[s.color].B + msg + color[s.color].E
	s.sugar.Errorf(msg, args)
}

func (s *Logger) Debugw(msg string, args ...interface{}) {
	msg = color[s.color].B + msg + color[s.color].E
	s.sugar.Debugw(msg, args...)
}

func (s *Logger) Infow(msg string, args ...interface{}) {
	msg = color[s.color].B + msg + color[s.color].E
	s.sugar.Infow(msg, args...)
}

func (s *Logger) Warnw(msg string, args ...interface{}) {
	msg = color[s.color].B + msg + color[s.color].E
	s.sugar.Warnw(msg, args...)
}

func (s *Logger) Errorw(msg string, args ...interface{}) {
	msg = color[s.color].B + msg + color[s.color].E
	s.sugar.Errorw(msg, args...)
}
