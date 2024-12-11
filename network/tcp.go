package network

import (
	"github.com/eesher/eegos/log"
	"github.com/eesher/eegos/util"

	"net"
	"runtime/debug"
	"time"
)

type TcpServer struct {
	TcpConn
	addr *net.TCPAddr
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

func (this *TcpServer) Start() {
	lis, err := net.ListenTCP("tcp", this.addr)

	log.Info("gateserver.Open: listening", this.addr)
	if err != nil {
		log.Error("gateserver.Open: net.ListenTCP: ", err)
		return
	}

	func(l *net.TCPListener) {
		//defer log.Debug("listen stop")
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

	defer func() {
		if err := recover(); err != nil {
			log.Error(err, string(debug.Stack()))
			s.Close()
		}
	}()

	this.handle.Connect(s.fd, s)
	go this.processInData(s)
}

func (this *TcpServer) processInData(s *tcpSession) {
	//defer log.Debug("processInData stop")
	for this.isOpen {
		select {
		case data, ok := <-s.inData:
			if !ok {
				continue
			}
			switch data.dType {
			case HEARTBEAT:
				if s.state != WORKING {
					break
				}
				go s.doWrite(data.head, HEARTBEAT_RET, []byte{})
				go this.handle.Heartbeat(s.fd, data.head)
			case DATA:
				go s.msgHandle(s.fd, data.head, data.body)
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
	session    *tcpSession
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
	//defer log.Debug("processInData stop")
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
			}
			go this.ticker.Reset(5 * time.Second)
		case <-s.cClose:
			this.Close()
			return
		}
	}
}

func (this *TcpClient) heartbeat() {
	//defer log.Debug("heartbeat stop")
	s := this.session
	for this.isOpen {
		<-this.ticker.C
		//log.Debug("heartbeat ticker")
		sessionID := this.msgCounter.GetNum()

		go s.doWrite(sessionID, HEARTBEAT, []byte{})

		select {
		case <-this.cHeartbeat:
			//log.Debug("heartbeat ret:", sid)

		case <-time.After(3 * time.Second):
			log.Warn("heartbeat Timed out", sessionID)
			this.ticker.Reset(5 * time.Second)
		}

	}
}

func (this *TcpClient) handleHeartbeatRet(fd uint16, sessionID uint16) {
	this.cHeartbeat <- sessionID
}

func (this *TcpClient) WriteData(s Session, buff []byte) uint16 {
	sessionID := this.msgCounter.GetNum()
	s.doWrite(sessionID, DATA, buff)
	return sessionID
}

func (this *TcpClient) Close() {
	//log.Debug("TcpClient Close()")
	this.isOpen = false
	this.ticker.Stop()
	close(this.cHeartbeat)
	this.TcpConn.Close(this.session)
}

type TcpConn struct {
	isOpen bool
	handle Handler
}

func (this *TcpConn) NewSession(conn net.Conn) *tcpSession {
	log.Debug("new connection from ", conn.RemoteAddr())
	session := CreateTcpSession(conn, this.handle.Message)
	session.Start()

	return session
}

func (this *TcpConn) Close(s *tcpSession) {
	//log.Debug("session close")
	this.handle.Close(s.fd)
	s.Release()
}

func (this *TcpConn) Write(s Session, sID uint16, buff []byte) {
	s.doWrite(sID, DATA, buff)
}
