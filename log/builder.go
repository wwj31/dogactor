package log

import (
	"fmt"
	"strconv"
	"time"
	"unsafe"

	"github.com/wwj31/godactor/log/colorized"
)

type Fields map[string]interface{}

func (field Fields) AddFiled(key string, data interface{}) {
	field[key] = data
}

type Builder struct {
	file  string
	line  int
	msg   string
	tag   string
	stack string

	buf      []byte
	param    []byte
	colorfun func(str string) string
	kv       Fields
	level    int32
}

func (b *Builder) String() string       { return *(*string)(unsafe.Pointer(&b.buf)) }
func (b *Builder) Len() int             { return len(b.buf) }
func (b *Builder) Cap() int             { return cap(b.buf) }
func (b *Builder) WriteString(s string) { b.buf = append(b.buf, s...) }
func (b *Builder) WriteBytes(s []byte)  { b.buf = append(b.buf, s...) }
func (b *Builder) Reset() {
	b.buf = b.buf[0:0]
	b.kv = make(Fields, 0)
	b.param = nil
	b.stack = ""
	b.colorfun = nil
	b.level = TAG_DEBUG_I
}
func (b *Builder) Bytes() []byte {
	time := time.Now().UTC().Format("2006-01-02 15:04:05")
	str := " [" + b.tag + "] " + ShortPath(b.file, 3) + ":" + strconv.Itoa(b.line)
	switch b.tag {
	case TAG_DEBUG:
		str = colorized.Blue(str)
	case TAG_INFO:
		str = colorized.Cyan(str)
	case TAG_WARN:
		str = colorized.Yellow(str)
	case TAG_ERROR:
		str = colorized.Red(str)
	case TAG_FATAL:
		str = colorized.Red(str)
	}
	if b.colorfun != nil {
		b.msg = b.colorfun(b.msg)
	}
	if len(b.msg) < 2 || b.msg[:2] != "\x1b[" {
		b.msg = colorized.White(b.msg)
	}
	str += " " + b.msg
	b.WriteString(fmt.Sprintf("%-120v", time+str))
	if len(b.param) > 0 {
		b.WriteString(fmt.Sprintf("param=%v", colorized.Green(string(b.param))))
	}
	b.WriteString("\n")
	b.WriteString(b.stack)
	return b.buf
}

func ShortPath(path string, need int) string {
	count := 0
	for i := len(path) - 1; i > 0; i-- {
		if path[i] == '/' {
			count++
		}
		if count == need {
			return path[i+1:]
		}
	}
	return path
}

func (b *Builder) KV(key string, value interface{}) *Builder {
	if b != nil {
		if err, ok := value.(error); ok {
			b.kv[key] = err.Error()
		} else if str, ok := value.(stringer); ok {
			b.kv[key] = str.String()
		} else {
			b.kv[key] = value
		}
	}

	return b
}

func (b *Builder) KVs(fields Fields) *Builder {
	if b != nil {
		for key, value := range fields {
			if err, ok := value.(error); ok {
				b.kv[key] = err.Error()
			} else if str, ok := value.(stringer); ok {
				b.kv[key] = str.String()
			} else {
				b.kv[key] = value
			}
		}
	}

	return b
}

func (b *Builder) Debug(msg string) {
	if b != nil {
		write(b, 0, TAG_DEBUG, msg)
	}
}
func (b *Builder) Info(msg string) {
	if b != nil {
		write(b, 0, TAG_INFO, msg)
	}
}
func (b *Builder) Warn(msg string) {
	if b != nil {
		write(b, 0, TAG_WARN, msg)
	}
}
func (b *Builder) Error(msg string) {
	if b != nil {
		write(b, 0, TAG_ERROR, msg)
	}
}
func (b *Builder) Fatal(msg string) {
	if b != nil {
		write(b, 0, TAG_FATAL, msg)
	}
}

func (b *Builder) DebugStack(stack int, msg string) {
	if b != nil {
		write(b, stack, TAG_DEBUG, msg)
	}
}
func (b *Builder) InfoStack(stack int, msg string) {
	if b != nil {
		write(b, stack, TAG_INFO, msg)
	}
}
func (b *Builder) WarnStack(stack int, msg string) {
	if b != nil {
		write(b, stack, TAG_WARN, msg)
	}
}
func (b *Builder) ErrorStack(stack int, msg string) {
	if b != nil {
		write(b, stack, TAG_ERROR, msg)
	}
}
func (b *Builder) FatalStack(stack int, msg string) {
	if b != nil {
		write(b, stack, TAG_FATAL, msg)
	}
}

func (b *Builder) Bule() *Builder {
	if b != nil {
		b.colorfun = colorized.Blue
	}
	return b
}
func (b *Builder) Yellow() *Builder {
	if b != nil {
		b.colorfun = colorized.Yellow
	}
	return b
}
func (b *Builder) Green() *Builder {
	if b != nil {
		b.colorfun = colorized.Green
	}
	return b
}
func (b *Builder) Magenta() *Builder {
	if b != nil {
		b.colorfun = colorized.Magenta
	}
	return b
}
func (b *Builder) Cyan() *Builder {
	if b != nil {
		b.colorfun = colorized.Cyan
	}
	return b
}
func (b *Builder) Gray() *Builder {
	if b != nil {
		b.colorfun = colorized.Gray
	}
	return b
}
func (b *Builder) White() *Builder {
	if b != nil {
		b.colorfun = colorized.White
	}
	return b
}
func (b *Builder) Red() *Builder {
	if b != nil {
		b.colorfun = colorized.Red
	}
	return b
}
