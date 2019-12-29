package log

import (
	"fmt"
	"io"
	olog "log"
)

type OriginLogger struct {
	w         io.Writer
	head      string
	callDepth int
}

func NewOriginLogger(w io.Writer) Logger {
	origin := OriginLogger{}
	origin.w = w
	origin.callDepth = 2

	return &origin
}

func (this *OriginLogger) VLog(v ...interface{}) error {
	olog.Output(this.callDepth, this.head+fmt.Sprintln(v...))
	return nil
}

func (this *OriginLogger) KVLog(v ...interface{}) error {
	olog.Output(this.callDepth, this.head+fmt.Sprintln(v...))
	return nil
}

func (this OriginLogger) SetDepth(depth int) Logger {
	this.callDepth = depth
	return &this
}

func (this *OriginLogger) SetFlags(flag int) {
	olog.SetFlags(flag)
}

func (this OriginLogger) WithHeader(keyvals ...interface{}) Logger {
	for i := 0; i < len(keyvals); i += 2 {
		this.head += "["
		this.head += keyvals[i+1].(string)
		this.head += "]: "
	}
	return &this
}
