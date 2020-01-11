package gate

import (
	"eegos/log"
	"eegos/util"

	//"bufio"
	"io"
	//	"net"
	"runtime/debug"
	"sync"
)

var writeLock = &sync.Mutex{}

var sessionCounter = util.Counter{Num: 0}

type Session struct {
	fd   uint16
	conn io.ReadWriteCloser
	//buf     *bufio.Writer
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
	//session.buf = bufio.NewWriter(conn)
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
	log.Debug("release session")
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
	pkg = make([]byte, 0, length+5)
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
		pkg = &Data{body: make([]byte, 0, pkgLen-5)}
		pkg.dType = data[2]
		pkg.head = uint16(data[3]) + uint16(data[4])<<8
		pkg.body = append(pkg.body, data[5:pkgLen]...)
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
		//log.Debug(idx, n, string(data))
		n = n + idx
		startIdx := 0
		for {
			pLen, pkg, _ := this.unpack(data[startIdx:], n)
			if pkg == nil {
				this.Reader(data, pLen)
				break
			} else {
				log.Debug(pLen, pkg.dType, pkg.head, string(pkg.body))
				this.inData <- pkg
				startIdx = pLen
				n = n - pLen
				if n == 0 {
					break
				}
			}
		}
	} else if err == io.EOF {
		return err
	} else {
		log.Error("read error:", err, n)
		return err
	}
	return nil
}

func (this *Session) NewReader() error {
	var b [5]byte
	if _, err := io.ReadFull(this.conn, b[:]); err != nil {
		return err
	}
	pkgLen := uint16(b[0]) + uint16(b[1])<<8
	pkg := &Data{}
	pkg.dType = b[2]
	pkg.head = uint16(b[3]) + uint16(b[4])<<8
	if pkgLen > 0 {
		pkg.body = make([]byte, pkgLen)
		if _, err := io.ReadFull(this.conn, pkg.body); err != nil {
			return err
		}
	}
	this.inData <- pkg
	return nil
}

func (this *Session) handleRead() {
	log.Debug("handleRead start")
	defer func() {
		log.Debug("connection close")
		this.conn.Close()
		this.cClose <- true
		if err := recover(); err != nil {
			log.Error("%s:\n%s\n", err, debug.Stack())
		}
	}()

	//buff := make([]byte, 1024)
	for {
		if this.state != WORKING {
			break
		}

		//buff = buff[:cap(buff)]
		//if err := this.Reader(buff, 0); err != nil {
		if err := this.NewReader(); err != nil {
			if err != io.EOF {
				log.Error(err)
			}
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
	log.Debug("handleRead stop")
}

func (this *Session) handleWrite() {
	log.Debug("handleWrite start")
	defer log.Debug("handleWrite stop")
	for {
		if this.state != WORKING {
			break
		}

		select {
		case pkg, ok := <-this.outData:
			if !ok {
				log.Error("pkg channel read error")
				break
			}
			//pLen, data, _ := this.unpack(pkg, 1024)
			//log.Debug("pkg write:", pLen, data.head)
			//pkg := this.pack(data.Head, data.dType, data.Body)
			this.conn.Write(pkg)
			//this.buf.Write(pkg)
			//this.buf.Flush()
		}
	}
}

//TODO make only one goroutine to write
func (this *Session) doWrite(head uint16, dType uint8, data []byte) {
	if this.state != WORKING {
		log.Error("session not working state=", this.state, "pkg not send", head)
		return
	}
	if data == nil {
		data = []byte{}
	}

	pkg := this.pack(head, dType, data)
	this.outData <- pkg
	//writeLock.Lock()
	//this.conn.Write(pkg)
	//this.buf.Write(pkg)
	//this.buf.Flush()
	//writeLock.Unlock()
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
