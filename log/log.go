package log

import (
	"bufio"
	"encoding/json"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const LogStack = 3

type LogHook func(t time.Time, level string, file string, line int, msg string)

var logHook LogHook
var logLevel int32
var buff *bufio.Writer
var exit int32
var wg sync.WaitGroup

func Stop() {
	if atomic.CompareAndSwapInt32(&exit, 0, 1) {
		wg.Wait()
	}
}
func Init(level int32, hook LogHook, dir, app string, appid int32) {
	logLevel = level
	logHook = hook

	baseLogPath := path.Join(dir, fmt.Sprintf("%v%v", app, appid))
	os.MkdirAll(baseLogPath, os.ModePerm)
	filename := fmt.Sprintf("%v%v", app, appid)
	outfile, err := rotatelogs.New(
		path.Join(baseLogPath, filename+".%Y%m%d%H%M.log"),
		rotatelogs.WithLinkName(path.Join(baseLogPath, filename)), // 生成软链，指向最新日志文件
		//rotatelogs.WithMaxAge(5*time.Second),     // 文件最大保存时间
		rotatelogs.WithRotationTime(2*time.Hour),  // 日志切割时间间隔
		rotatelogs.WithRotationSize(1024*1024*50), // 日志切割大小 50M一个文件
	)

	log.SetOutput(io.MultiWriter(os.Stdout, outfile))
	buff = bufio.NewWriter(log.Writer())
	if err != nil {
		KV("err", errors.WithStack(err)).Error("config local file system logger error.")
	}
}

func WhereCall(depth int, shortPath int) (string, int) {
	if _, file, line, ok := runtime.Caller(depth); ok {
		if shortPath > 0 {
			file = ShortPath(file, shortPath)
		}
		return file, line
	}
	return "???", 0
}

// 输出堆栈信息, skip表示需要跳过的帧数
func Backtrace(skip, needline int) string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	buf = buf[:n]
	s := string(buf)
	lines := skip
	count := 0
	begin := strings.IndexFunc(s, func(c rune) bool {
		if c != '\n' {
			return false
		}
		count++
		return count == lines
	})

	count = 0
	end := strings.IndexFunc(s, func(c rune) bool {
		if c != '\n' {
			return false
		}
		count++
		return count == lines+needline+2
	})
	return s[begin+1 : end+1]
}

const (
	TAG_DEBUG = "D"
	TAG_INFO  = "I"
	TAG_WARN  = "W"
	TAG_ERROR = "E"
	TAG_FATAL = "F"
)

const (
	TAG_DEBUG_I = iota
	TAG_INFO_I
	TAG_WARN_I
	TAG_ERROR_I
	TAG_FATAL_I
)

var Levels = map[string]int32{
	TAG_DEBUG: TAG_DEBUG_I,
	TAG_INFO:  TAG_INFO_I,
	TAG_WARN:  TAG_WARN_I,
	TAG_ERROR: TAG_ERROR_I,
	TAG_FATAL: TAG_FATAL_I,
}

var msgs = make(chan *Builder, 10000)
var msgPool = &sync.Pool{
	New: func() interface{} {
		return &Builder{buf: make([]byte, 0, 128), kv: make(Fields, 0)}
	},
}

func init() {
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(200 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				if buff != nil {
					buff.Flush()
				}
			case msg := <-msgs:
				if buff != nil {
					buff.Write(msg.Bytes())
					msg.Reset()
					msgPool.Put(msg)
					if logHook != nil {
						logHook(time.Now(), msg.tag, msg.file, msg.line, msg.msg)
					}
				}
			}

			if len(msgs) > 0 {
				continue
			} else if atomic.LoadInt32(&exit) == 1 {
				buff.Flush()
				return
			}
		}
	}()
}

func write(sb *Builder, stack int, tag string, fm string, v ...interface{}) {
	if Levels[tag] < logLevel {
		return
	}
	if sb != nil && Levels[tag] < sb.level {
		return
	}

	if sb == nil {
		sb = msgPool.Get().(*Builder)
	}

	sb.file, sb.line = WhereCall(LogStack, -1)
	sb.msg = fmt.Sprintf(fm, v...)
	sb.tag = tag

	if len(sb.kv) > 0 {
		sb.param, _ = json.Marshal(sb.kv)
	}

	if stack > 0 {
		sb.stack = Backtrace(8, stack)
	}

	msgs <- sb
}

type stringer interface {
	String() string
}

func KV(key string, value interface{}) *Builder {
	sb := msgPool.Get().(*Builder)
	if err, ok := value.(error); ok {
		sb.kv[key] = err.Error()
	} else if str, ok := value.(stringer); ok {
		sb.kv[key] = str.String()
	} else {
		sb.kv[key] = value
	}
	return sb
}

func KVs(fields Fields) *Builder {
	sb := msgPool.Get().(*Builder)
	if fields == nil {
		return sb
	}

	for key, value := range fields {
		if err, ok := value.(error); ok {
			sb.kv[key] = err.Error()
		} else if str, ok := value.(stringer); ok {
			sb.kv[key] = str.String()
		} else {
			sb.kv[key] = value
		}
	}
	return sb
}

func Debug(msg string) { write(nil, 0, TAG_DEBUG, msg) }
func Info(msg string)  { write(nil, 0, TAG_INFO, msg) }
func Warn(msg string)  { write(nil, 0, TAG_WARN, msg) }
func Error(msg string) { write(nil, 0, TAG_ERROR, msg) }
func Fatal(msg string) { write(nil, 0, TAG_FATAL, msg) }

func DebugStack(stack int, msg string) { write(nil, stack, TAG_DEBUG, msg) }
func InfoStack(stack int, msg string)  { write(nil, stack, TAG_INFO, msg) }
func WarnStack(stack int, msg string)  { write(nil, stack, TAG_WARN, msg) }
func ErrorStack(stack int, msg string) { write(nil, stack, TAG_ERROR, msg) }
func FatalStack(stack int, msg string) { write(nil, stack, TAG_FATAL, msg) }
