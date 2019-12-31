package gate

import (
	"eegos/log"
	"eegos/util"

	"bufio"
	"io"
	"net"
	"runtime/debug"
)

var sessionCounter = util.Counter{Num: 0}

type Session struct {
	fd      uint16
	conn    io.ReadWriteCloser
	buf     *bufio.Writer
	inData  chan *Data
	outData chan *Data
	cClose  chan bool
	state   int
}

func CreateSession(conn *net.TCPConn) *Session {
	session := new(Session)
	session.fd = sessionCounter.GetNum()
	session.conn = conn
	session.buf = bufio.NewWriter(conn)
	session.inData = make(chan *Data, 1)
	session.outData = make(chan *Data, 1)
	session.cClose = make(chan bool)
	session.state = NEW_CONNECTION

	return session
}

func (this *Session) Start() {
	this.state = WORKING
	go this.handleRead()
	//go this.handleWrite()
}

func (this *Session) Close() {
	this.state = CLOSING
}

func (this *Session) Release() {
	close(this.inData)
	close(this.outData)
	close(this.cClose)
	this.state = CLOSED
}

func (this *Session) pack(head uint16, dType uint8, body []byte) (pkg []byte) {
	length := 5 + len(body)
	pkg = make([]byte, 0, length)
	pkg = append(pkg, uint8(length), uint8(length>>8), dType, uint8(head), uint8(head>>8))
	//pkg = append(pkg, uint8(length>>8))
	//pkg = append(pkg, dType)
	//pkg = append(pkg, uint8(head), uint8(head>>8))
	//pkg = append(pkg, uint8(session_id>>8))
	pkg = append(pkg, body...)

	/*
		if body != nil {
			//var protoPkg []byte
			//protoPkg, err = this.marshal(v)
			length += len(body)
			pkg[0] = uint8(length)
			pkg[1] = uint8(length >> 8)
			pkg = append(pkg, body...)
		}
	*/
	return pkg
}

func (this *Session) unpack(data []byte, length int) (pkgLen int, pkg *Data, err error) {
	pkgLen = int(data[0]) + int(data[1])<<8
	if pkgLen <= length {
		pkg = &Data{Body: make([]byte, 0, pkgLen-5)}
		pkg.dType = data[2]
		pkg.Head = uint16(data[3]) + uint16(data[4])<<8
		pkg.Body = append(pkg.Body, data[5:pkgLen]...)
		return
	} else {
		pkgLen = length
	}

	return pkgLen, nil, nil
}

func (this *Session) Reader(data []byte, idx int) error {
	if n, err := this.conn.Read(data[idx:]); err == nil {
		if n == 0 {
			return nil
		}
		n = n + idx
		startIdx := 0
		for {
			pLen, pkg, _ := this.unpack(data[startIdx:], n)
			if pkg == nil {
				this.Reader(data, pLen)
				break
			} else {
				this.inData <- pkg
				startIdx = pLen
				n = n - pLen
				if n == 0 {
					break
				}
			}
		}
	} else {
		log.Error("read error:", err, n)
		return err
	}
	return nil
}

func (this *Session) handleRead() {
	defer func() {
		log.Debug("connection close")
		this.conn.Close()
		this.cClose <- true
		if err := recover(); err != nil {
			log.Error("%s:\n%s\n", err, debug.Stack())
		}
	}()

	buff := make([]byte, 1024)
	for {
		if this.state != WORKING {
			break
		}

		if err := this.Reader(buff, 0); err != nil {
			log.Error(err)
			break
		}
	}

	/*
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		select {
		case len(this.newData) == 0:
			return
		case <-ticker.C:
			log.Warnf("lose %d package, fd:%d", len(this.newData), this.fd)
			return
		}
	*/
}

func (this *Session) handleWrite() {
	for {
		if this.state != WORKING {
			break
		}

		select {
		case data, ok := <-this.outData:
			if !ok {
				continue
			}
			pkg := this.pack(data.Head, data.dType, data.Body)
			this.buf.Write(pkg)
			this.buf.Flush()
		}
	}
	log.Debug("handleWrite stopped")
}

func (this *Session) Write(head uint16, dType uint8, data []byte) {
	if this.state != WORKING {
		log.Error("session not working state=", this.state, "pkg not send", head)
		return
	}
	if data == nil {
		data = []byte{}
	}

	pkg := this.pack(head, dType, data)
	this.buf.Write(pkg)
	this.buf.Flush()
}

/*
func (this *Session) DoWrite(session_id uint16, dType uint8, data []interface{}) {
	ret_data, _ := this.Pack(session_id, dType, data)
	//log.Println("send data:", string(ret_data), len(ret_data))
	this.buf.Write(ret_data)
	this.buf.Flush()
}

func (this *Session) HandleWrite(session_id uint16, data []interface{}) {
	this.DoWrite(session_id, packData, data)
}

func (this *Session) HandleWriteTest(session_id uint16, data []interface{}) {
	ret_data, _ := this.Pack(session_id, packData, data)
	//log.Println("send data:", string(ret_data[4:]))
	this.buf.Write(ret_data[:5])
	this.buf.Flush()
	this.buf.Write(ret_data[5:])
	this.buf.Flush()
}
*/
