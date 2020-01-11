package gate

import (
	"eegos/log"
	"eegos/util"

	"net"
	"time"
)

type Handler interface {
	Connect(uint16, *Session)
	Message(uint16, uint16, []byte)
	Heartbeat(uint16, uint16)
	Close(uint16)
}

type TcpServer struct {
	TcpConn
	addr *net.TCPAddr
	//lis *net.TCPListener
}

func NewTcpServer(handle Handler, addr string) *TcpServer {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		log.Error("gateserver.Open: net.ResolveTCPAddr: ", err)
		return nil
	}
	newServer := &TcpServer{TcpConn{isOpen: true, handle: handle},
		tcpAddr,
	}
	return newServer
}

/*
func (this *TcpServer) Open(addr string) {
}
*/

func (this *TcpServer) Start() {
	lis, err := net.ListenTCP("tcp", this.addr)

	log.Info("gateserver.Open: listening")
	if err != nil {
		log.Error("gateserver.Open: net.ListenTCP: ", err)
		return
	}

	go func(l *net.TCPListener) {
		defer log.Debug("listen stop")
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Error("gateserver.Open: lis.Accept: ", err)
				return
			}
			//conn.SetKeepAlive(true)
			//conn.SetKeepAlivePeriod(5 * time.Second)
			go this.handleNewConn(conn)
		}
	}(lis)
}

func (this *TcpServer) handleNewConn(conn net.Conn) {
	s := this.NewSession(conn)
	this.handle.Connect(s.fd, s)
	go this.processInData(s)
}

func (this *TcpServer) processInData(s *Session) {
	defer log.Debug("processInData stop")
	for this.isOpen {
		select {
		case data, ok := <-s.inData:
			//log.Debug("server in data", data.dType, data.head, ok)
			if !ok {
				continue
			}
			switch data.dType {
			case HEARTBEAT:
				go s.doWrite(data.head, HEARTBEAT_RET, []byte{})
				//go s.Write(data.Head, HEARTBEAT_RET, []byte{})
				go this.handle.Heartbeat(s.fd, data.head)
			case DATA:
				go s.msgHandle(s.fd, data.head, data.body)
				//go this.handle.Message(s.fd, data.head, data.body)
			}
		case <-s.cClose:
			this.Close(s)
			return
		}
	}
}

type TcpClient struct {
	TcpConn
	msgCounter *util.Counter
	cHeartbeat chan uint16
	ticker     *time.Timer
	session    *Session
}

func NewTcpClient(handle Handler) *TcpClient {
	newClient := &TcpClient{TcpConn{isOpen: true, handle: handle},
		&util.Counter{Num: 0},
		make(chan uint16),
		time.NewTimer(5 * time.Second),
		nil}
	return newClient
}

func (this *TcpClient) Dial(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Error("net.Dial: ", err)
		return
	}

	s := this.NewSession(conn)
	this.handle.Connect(s.fd, s)
	this.session = s
	go this.processInData()
	go this.heartbeat()
}

func (this *TcpClient) GetSessionID() uint16 {
	return this.msgCounter.GetNum()
}

func (this *TcpClient) processInData() {
	defer log.Debug("processInData stop")
	//defer this.Close(s.fd)
	s := this.session
	for this.isOpen {
		select {
		case data, ok := <-s.inData:
			if !ok {
				continue
			}
			switch data.dType {
			case HEARTBEAT_RET:
				go this.handleHeartbeatRet(s.fd, data.head)
			case DATA:
				go s.msgHandle(s.fd, data.head, data.body)
				//go this.handle.Message(s.fd, data.head, data.body)
			}
			go this.ticker.Reset(5 * time.Second)
		case <-s.cClose:
			this.Close()
			return
		}
	}
}

func (this *TcpClient) heartbeat() {
	defer log.Debug("heartbeat stop")
	s := this.session
	for this.isOpen {
		<-this.ticker.C
		log.Debug("heartbeat ticker")
		sessionID := this.msgCounter.GetNum()

		go s.doWrite(sessionID, HEARTBEAT, []byte{})
		//go s.Write(sessionID, HEARTBEAT, []byte{})

		select {
		case sid := <-this.cHeartbeat:
			log.Debug("heartbeat ret:", sid)

		case <-time.After(3 * time.Second):
			log.Debug("heartbeat Timed out", sessionID)
			this.ticker.Reset(5 * time.Second)
		}

	}
}

func (this *TcpClient) handleHeartbeatRet(fd uint16, sessionID uint16) {
	this.cHeartbeat <- sessionID
}

func (this *TcpClient) WriteData(s *Session, buff []byte) uint16 {
	sessionID := this.msgCounter.GetNum()
	s.doWrite(sessionID, DATA, buff)
	return sessionID
	//pack := &Data{Head: sessionID, dType: DATA, Body: buff}
}

func (this *TcpClient) Close() {
	log.Debug("TcpClient Close()")
	this.isOpen = false
	this.ticker.Stop()
	close(this.cHeartbeat)
	this.TcpConn.Close(this.session)
}

type TcpConn struct {
	isOpen bool
	handle Handler
}

func (this *TcpConn) NewSession(conn net.Conn) *Session {
	log.Debug("new connection from ", conn.RemoteAddr())
	session := CreateSession(conn, this.handle.Message)
	session.Start()

	//go this.processOutData(session)
	return session
}

func (this *TcpConn) Close(s *Session) {
	//this.isOpen = false
	log.Debug("session close")
	this.handle.Close(s.fd)
	//this.Close(s.fd)
	s.Release()
}

/*
func (this *TcpConn) processOutData(s *Session) {
	defer log.Debug("processOutData close")
	for this.isOpen {
		select {
		case data, ok := <-s.outData:
			if !ok {
				continue
			}
			if data.dType == PKG_TYPE {
				data.dType = DATA
			}
			log.Debug("out data", data.Head)
			go s.Write(data.Head, data.dType, data.Body)

		case <-s.cClose:
			this.Close(s)
			return
		}
	}
}
*/

func (this *TcpConn) Write(s *Session, sID uint16, buff []byte) {
	s.doWrite(sID, DATA, buff)
}
