package rpc

import (
	"bufio"
	"io"
	"log"
	"net"
	"runtime/debug"
)

const (
	PackRegister = iota
	HeartBeat
	PackData
)

type Session struct {
	conn    io.ReadWriteCloser
	buf     *bufio.Writer
	newData chan *Data
	cClose  chan bool
}

type CallInfo struct {
	session_id uint
	args       []interface{}
	cCallBack  chan []interface{}
}

func CreateSession(conn net.Conn) *Session {
	session := new(Session)
	session.conn = conn
	session.buf = bufio.NewWriter(conn)
	session.newData = make(chan *Data, 1)
	session.cClose = make(chan bool)

	go session.HandleRead()
	return session
}

func (this *Session) Reader(data []byte, idx int) error {
	if n, err := this.conn.Read(data[idx:]); err == nil {
		if n == 0 {
			return nil
		}
		n = n + idx
		startIdx := 0
		for {
			pLen, pkg, _ := Unpack(data[startIdx:], n)
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
		log.Println("read error:", err, n)
		return err
	}
	return nil
}

func (this *Session) HandleRead() {
	defer func() {
		log.Println("connection close")
		this.conn.Close()
		this.cClose <- true
		if err := recover(); err != nil {
			log.Printf("%s:\n%s\n", err, debug.Stack())
		}
	}()

	buff := make([]byte, 128)
	for {
		if err := this.Reader(buff, 0); err != nil {
			log.Println(err)
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

func (this *Session) HandleWrite(session_id uint, data []interface{}) {
	ret_data, _ := Pack(session_id, PackData, data)
	//log.Println("send data:", string(ret_data))
	this.buf.Write(ret_data)
	this.buf.Flush()
}

func (this *Session) HandleWriteTest(session_id uint, data []interface{}) {
	ret_data, _ := Pack(session_id, PackData, data)
	//log.Println("send data:", string(ret_data[4:]))
	this.buf.Write(ret_data[:5])
	this.buf.Flush()
	this.buf.Write(ret_data[5:])
	this.buf.Flush()
}
