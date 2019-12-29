package gate

import (
	"eegos/log"

	"bufio"
	"io"
	"net"
	"runtime/debug"
)

type Session struct {
	conn    io.ReadWriteCloser
	buf     *bufio.Writer
	newData chan *Data
	cClose  chan bool
}

type Data struct {
	data_type uint8
	head      uint16
	body      []byte
}

func CreateSession(conn net.TCPConn) *Session {
	session := new(Session)
	session.conn = conn
	session.buf = bufio.NewWriter(conn)
	session.newData = make(chan *Data, 1)
	session.cClose = make(chan bool)

	go session.HandleRead()
	return session
}

func (this *Session) Pack(session_id uint16, data_type uint8, body []byte) (pkg []byte) {
	length := 5
	pkg = []byte{}
	pkg = append(pkg, uint8(length))
	pkg = append(pkg, uint8(length>>8))
	pkg = append(pkg, data_type)
	pkg = append(pkg, uint8(session_id))
	pkg = append(pkg, uint8(session_id>>8))
	if v != nil {
		//var protoPkg []byte
		//protoPkg, err = this.marshal(v)
		length += len(body)
		pkg[0] = uint8(length)
		pkg[1] = uint8(length >> 8)
		pkg = append(pkg, body...)
	}
	return pkg
}

func (this *Session) Unpack(data []byte, length int) (pkgLen int, pkg *Data, err error) {
	pkgLen = int(data[0]) + int(data[1])<<8
	if pkgLen <= length {
		pkg = &Data{}
		pkg.data_type = data[2]
		pkg.head = uint16(data[3]) + uint16(data[4])<<8
		pkg.body = data[5:pkgLen]
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
			pLen, pkg, _ := this.Unpack(data[startIdx:], n)
			if pkg == nil {
				this.Reader(data, pLen)
				break
			} else {
				this.newData <- pkg
				startIdx = pLen
				n = n - pLen
				if n == 0 {
					break
				}
			}
		}
	} else {
		log.error("read error:", err, n)
		return err
	}
	return nil
}

func (this *Session) HandleRead() {
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
		if err := this.Reader(buff, 0); err != nil {
			log.Error(err)
			break
		}
	}
	/*
		for {
			log.Println("try read")
			data := make([]byte, 128)
			if n, err := this.conn.Read(data); err == nil {
				idx := 0
				for {
					info := Data{}
					idx, info.head, info.body, _ = Unpack(data[idx:], n-idx)
					this.newData <- info
					if idx == 0 {
						break
					}
					log.Println("pkg connect")
				}

			} else {
				log.Println("read error:", err, n)
				break
			}
			//log.Println("read")
		}
	*/
}

func (this *Session) Write(session_id uint16, data_type uint8, data []byte) {
	pkg := this.Pack(session_id, data_type, data)
	this.buf.Write(pkg)
	this.buf.Flush()
}

/*
func (this *Session) DoWrite(session_id uint16, data_type uint8, data []interface{}) {
	ret_data, _ := this.Pack(session_id, data_type, data)
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
