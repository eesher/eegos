package log

import (
	"fmt"
	"os"
)

type Logger interface {
	VLog(keyvals ...interface{}) error
	KVLog(keyvals ...interface{}) error
	WithHeader(keyvals ...interface{}) Logger
	SetDepth(depth int) Logger
	SetFlags(flag int)
}

var defaultLogger = NewOriginLogger(os.Stdout)

const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

const (
	NOLOG = 0
	DEBUG = 1
	INFO  = 2
	WARN  = 3
	ERROR = 4
)

var (
	LEVEL_NAME map[int]string
	LOG_LEVEL  int
	OPEN_STACK bool
)

func doLog(level int, useV bool, v ...interface{}) {
	if LOG_LEVEL > level {
		return
	}
	tmpl := defaultLogger.WithHeader("level", LEVEL_NAME[level])
	tmpl = tmpl.SetDepth(4)
	if useV {
		tmpl.VLog(v...)
	} else {
		tmpl.KVLog(v...)
	}
}

func doFormatLog(level int, format string, v ...interface{}) {
	if LOG_LEVEL > level {
		return
	}
	tmpl := defaultLogger.WithHeader("level", LEVEL_NAME[level])
	tmpl = tmpl.SetDepth(4)
	tmpl.VLog(fmt.Sprintf(format, v...))
}

func Log(v ...interface{}) {
	doLog(DEBUG, true, v...)
}

func Logf(format string, v ...interface{}) {
	doFormatLog(DEBUG, format, v...)
}

func Debug(v ...interface{}) {
	doLog(DEBUG, false, v...)
}

func Debugf(format string, v ...interface{}) {
	doFormatLog(DEBUG, format, v...)
}

func Info(v ...interface{}) {
	doLog(INFO, false, v...)
}

func Infof(format string, v ...interface{}) {
	doFormatLog(INFO, format, v...)
}

func Warn(v ...interface{}) {
	doLog(WARN, false, v...)
}

func Warnf(format string, v ...interface{}) {
	doFormatLog(WARN, format, v...)
}

func Error(v ...interface{}) {
	doLog(ERROR, false, v...)
}

func Errorf(format string, v ...interface{}) {
	doFormatLog(ERROR, format, v...)
}

func WithHeader(keyvals ...interface{}) Logger {
	return defaultLogger.WithHeader(keyvals...)
}

func SetFlags(flag int) {
	defaultLogger.SetFlags(flag)
}

func OpenStack() {
	OPEN_STACK = true
}

func SetLevel(level int) {
	LOG_LEVEL = level
}

func init() {
	LOG_LEVEL = 0
	OPEN_STACK = false
	LEVEL_NAME = make(map[int]string)
	LEVEL_NAME[NOLOG] = "NOLOG"
	LEVEL_NAME[DEBUG] = "DEBUG"
	LEVEL_NAME[INFO] = "INFO"
	LEVEL_NAME[WARN] = "WARN"
	LEVEL_NAME[ERROR] = "ERROR"
}
