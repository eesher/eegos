package network

import (
	"github.com/eesher/eegos/log"
	"github.com/eesher/eegos/util"

	"io"
	"runtime/debug"
	"sync"
)

var writeLock = &sync.Mutex{}

var sessionCounter = util.Counter{Num: 0}

type Session struct {
	fd        uint16
	conn      io.ReadWriteCloser
	inData    chan *Data
	outData   chan []byte
	cClose    chan bool
	state     int
	msgHandle func(uint16, uint16, []byte)
}

func CreateSession(conn io.ReadWriteCloser, msgHandle func(uint16, uint16, []byte)) *Session {
	session := new(Session)
	session.fd = sessionCounter.GetNum()
	session.conn = conn
	session.inData = make(chan *Data, 1)
	session.outData = make(chan []byte, 1)
	session.cClose = make(chan bool)
	session.state = NEW_CONNECTION
	session.msgHandle = msgHandle

	return session
}

func (this *Session) Start() {
	this.state = WORKING
	go this.handleRead()
	go this.handleWrite()
}

func (this *Session) Close() {
	this.state = CLOSING
}

func (this *Session) Release() {
	//log.Debug("release session")
	close(this.inData)
	close(this.outData)
	close(this.cClose)
	this.state = CLOSED
}

func (this *Session) Forward(msgHandle func(uint16, uint16, []byte)) {
	this.msgHandle = msgHandle
}

func (this *Session) pack(head uint16, dType uint8, body []byte) (pkg []byte) {
	length := len(body)
	if length > 65535 {
		log.Warn("package too big, limit 65535!!!!", length)
		length = 0
		body = []byte{}
	}
	pkg = make([]byte, 0, length+5)
	pkg = append(pkg, uint8(length), uint8(length>>8), dType, uint8(head), uint8(head>>8))
	pkg = append(pkg, body...)
	return pkg
}

func (this *Session) Reader() (err error) {
	var b [5]byte
	if _, err = io.ReadFull(this.conn, b[:]); err != nil {
		return
	}
	pkgLen := uint16(b[0]) + uint16(b[1])<<8
	pkg := &Data{}
	pkg.dType = b[2]
	pkg.head = uint16(b[3]) + uint16(b[4])<<8
	if pkgLen > 0 {
		pkg.body = make([]byte, pkgLen)
		if _, err = io.ReadFull(this.conn, pkg.body); err != nil {
			return
		}
	}
	this.inData <- pkg
	return nil
}

func (this *Session) handleRead() {
	//log.Debug("handleRead start")
	defer func() {
		//log.Debug("connection close")
		this.conn.Close()
		this.cClose <- true
		if err := recover(); err != nil {
			log.Error(err, string(debug.Stack()))
		}
	}()

	for {
		if this.state != WORKING {
			break
		}

		if err := this.Reader(); err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				log.Error(err)
			}
			break
		}
	}

	//log.Debug("handleRead stop")
}

func (this *Session) handleWrite() {
	//log.Debug("handleWrite start")
	//defer log.Debug("handleWrite stop")
	for {
		if this.state != WORKING {
			break
		}

		select {
		case pkg, ok := <-this.outData:
			if ok {
				this.conn.Write(pkg)
			}
		}
	}
	this.conn.Close()
}

func (this *Session) doWrite(head uint16, dType uint8, data []byte) {
	if this.state != WORKING {
		log.Error("session not working state=", this.state, "pkg not send", head, dType)
		return
	}
	if data == nil {
		data = []byte{}
	}

	pkg := this.pack(head, dType, data)
	this.outData <- pkg
}
