package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type logfmtEncoder struct {
	*Encoder
	buf bytes.Buffer
}

func (l *logfmtEncoder) Reset() {
	l.Encoder.Reset()
	l.buf.Reset()
}

var logfmtEncoderPool = sync.Pool{
	New: func() interface{} {
		var enc logfmtEncoder
		enc.Encoder = NewEncoder(&enc.buf)
		return &enc
	},
}

type InfluxdbLogger struct {
	mu          sync.Mutex // ensures atomic writes; protects the following fields
	w           io.Writer
	measurement string
	tagSet      bytes.Buffer
	buf         bytes.Buffer
	callDepth   int
	flag        int
}

// NewLogfmtLogger returns a logger that encodes keyvals to the Writer in
// logfmt format. Each log event produces no more than one call to w.Write.
// The passed Writer must be safe for concurrent use by multiple goroutines if
// the returned Logger will be used concurrently.
/*
func NewInfluxdbLogger(w io.Writer) Logger {
	influxdb := InfluxdbLogger{}
	influxdb.w = w
	arg := os.Args[0]
	idx := strings.LastIndex(arg, "/")
	influxdb.measurement = arg[idx+1:]

	return &influxdb
}
*/

func NewInfluxdbLogger(w io.Writer, measurement string, kv ...string) Logger {
	influxdb := InfluxdbLogger{}
	influxdb.w = w
	influxdb.callDepth = 2

	if measurement == "" {
		arg := os.Args[0]
		idx := strings.LastIndex(arg, "/")
		influxdb.measurement = arg[idx+1:]
	} else {
		influxdb.measurement = measurement
	}

	if len(kv) > 0 {
		influxdb.tagSet.WriteString(",")
		for i := 0; i < len(kv); i += 2 {
			key, value := kv[i], kv[i+1]
			influxdb.tagSet.WriteString(key)
			influxdb.tagSet.WriteString("=")
			influxdb.tagSet.WriteString(value)
			if i < len(kv)-2 {
				influxdb.tagSet.WriteString(",")
			}
		}
	}

	defaultLogger = &influxdb
	return &influxdb
}

func (l *InfluxdbLogger) VLog(v ...interface{}) error {
	caller := l.getCaller()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf.Reset()
	l.buf.WriteString(l.measurement)

	//l.tagSet.WriteString(",")
	l.buf.Write(l.tagSet.Bytes())
	l.buf.Write([]byte(" msg="))

	l.buf.WriteString("\"")
	for i := 0; i < len(v); i++ {
		value := v[i]
		l.buf.WriteString(fmt.Sprint(value))
		if i < len(v)-1 {
			l.buf.WriteString(" ")
		}
	}
	l.buf.WriteString("\"")

	l.buf.WriteString(",")
	l.buf.WriteString(caller)
	l.buf.WriteString(" ")

	l.buf.WriteString(fmt.Sprint(time.Now().UnixNano()))

	l.buf.WriteString("\n")

	if _, err := l.w.Write(l.buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func (l *InfluxdbLogger) KVLog(keyvals ...interface{}) error {
	enc := logfmtEncoderPool.Get().(*logfmtEncoder)
	enc.Reset()
	defer logfmtEncoderPool.Put(enc)

	enc.buf.WriteString(l.measurement)
	//enc.buf.WriteString(",")
	enc.buf.Write(l.tagSet.Bytes())

	enc.buf.Write([]byte(" "))

	if err := enc.EncodeKeyvalsWithQuoted(keyvals...); err != nil {
		return err
	}
	enc.buf.WriteString(",")
	enc.buf.WriteString(l.getCaller())
	enc.buf.WriteString(" ")

	//enc.buf.Write([]byte(" "))
	enc.buf.WriteString(fmt.Sprint(time.Now().UnixNano()))

	// Add newline to the end of the buffer
	if err := enc.EndRecord(); err != nil {
		return err
	}

	// The Logger interface requires implementations to be safe for concurrent
	// use by multiple goroutines. For this implementation that means making
	// only one call to l.w.Write() for each call to Log.
	if _, err := l.w.Write(enc.buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func (l InfluxdbLogger) SetDepth(depth int) Logger {
	l.callDepth = depth
	return &l
}

func (l *InfluxdbLogger) SetFlags(flag int) {
	l.flag = flag
}

/*
func (l *InfluxdbLogger) WithHeader(keyvals ...interface{}) error {
	enc := logfmtEncoderPool.Get().(*logfmtEncoder)
	enc.Reset()
	defer logfmtEncoderPool.Put(enc)

	if l.tagFlag {
		l.tagSet.Write([]byte(","))
	}

	if err := enc.EncodeKeyvals(keyvals...); err != nil {
		return err
	}
	l.tagSet.Write(enc.buf.Bytes())
	l.tagFlag = true
	return nil
}
*/

func (l InfluxdbLogger) WithHeader(keyvals ...interface{}) Logger {
	enc := logfmtEncoderPool.Get().(*logfmtEncoder)
	enc.Reset()
	defer logfmtEncoderPool.Put(enc)

	if err := enc.EncodeKeyvals(keyvals...); err != nil {
		fmt.Println(err)
		return nil
	}
	l.tagSet.Write([]byte(","))
	l.tagSet.Write(enc.buf.Bytes())
	return &l
}

/*
func (l *InfluxdbLogger) SetHeader(head string) {
	l.header.Reset()
	l.header.WriteString(head)
}
*/

func (l *InfluxdbLogger) getCaller() string {
	_, file, line, ok := runtime.Caller(l.callDepth)
	if !ok {
		file = "???"
		line = 0
	}

	if l.flag&Lshortfile != 0 {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
	}
	return "caller=\"" + file + ":" + fmt.Sprint(line) + "\""
}
