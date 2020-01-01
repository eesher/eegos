package gate

import (
	"eegos/log"
	"eegos/util"

	"net"
	"time"
)

type Handler interface {
	Connect(uint16, chan *Data)
	Message(uint16, uint16, []byte)
	Heartbeat(uint16, uint16)
	Close(uint16)
}

type TcpServer struct {
	isOpen bool
	handle Handler
}

func NewTcpServer(handle Handler) *TcpServer {
	return &TcpServer{isOpen: true, handle: handle}
}

func (this *TcpServer) Open(addr string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		log.Error("gateserver.Open: net.ResolveTCPAddr: ", err)
		return
	}
	lis, err := net.ListenTCP("tcp", tcpAddr)

	log.Info("gateserver.Open: listening")
	if err != nil {
		log.Error("gateserver.Open: net.ListenTCP: ", err)
		return
	}

	go func() {
		for this.isOpen {
			conn, err := lis.AcceptTCP()
			if err != nil {
				log.Error("gateserver.Open: lis.Accept: ", err)
				return
			}
			conn.SetKeepAlive(true)
			conn.SetKeepAlivePeriod(5 * time.Second)
			go this.NewSession(conn)
		}
	}()
}

func (this *TcpServer) NewSession(conn *net.TCPConn) {
	log.Debug("new connection from ", conn.RemoteAddr())
	session := CreateSession(conn)
	//handle new connection
	this.handle.Connect(session.fd, session.outData)
	session.Start()

	go this.processData(session)
}

func (this *TcpServer) processData(s *Session) {
	for {
		select {
		case data, ok := <-s.inData:
			if !ok {
				continue
			}
			switch data.dType {
			case HEARTBEAT:
				go s.Write(data.Head, HEARTBEAT_RET, []byte{})
				go this.handle.Heartbeat(s.fd, data.Head)
			//case HEARTBEAT_RET:
			//go this.handle.HeartBeat(s.fd, data.Head)
			case DATA:
				go this.handle.Message(s.fd, data.Head, data.Body)
			}
		case data, ok := <-s.outData:
			if !ok {
				continue
			}
			go s.Write(data.Head, DATA, data.Body)

		case c := <-s.cClose:
			log.Debug("session close", c)
			this.handle.Close(s.fd)
			s.Release()
			return
		}
	}
}

type TcpClient struct {
	isOpen     bool
	handle     Handler
	cHeartbeat chan uint16
	ticker     *time.Ticker
	MsgCounter util.Counter
}

func NewTcpClient(handle Handler) *TcpClient {
	newClient := &TcpClient{isOpen: true, handle: handle}
	newClient.cHeartbeat = make(chan uint16)
	newClient.ticker = time.NewTicker(5 * time.Second)
	newClient.MsgCounter = util.Counter{Num: 0}
	return newClient
}

func (this *TcpClient) Dial(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Error("net.Dial: ", err)
		return
	}

	session := CreateSession(conn)
	this.handle.Connect(session.fd, session.outData)
	session.Start()
	go this.processData(session)
	go this.heartbeat(session)
}

func (this *TcpClient) processData(s *Session) {
	for {
		select {
		case data, ok := <-s.inData:
			if !ok {
				continue
			}
			switch data.dType {
			case HEARTBEAT:
				go s.Write(data.Head, HEARTBEAT_RET, []byte{})
				go this.handle.Heartbeat(s.fd, data.Head)
			case HEARTBEAT_RET:
				go this.handleHeartbeatRet(s.fd, data.Head)
			case DATA:
				go this.handle.Message(s.fd, data.Head, data.Body)
			}
		case data, ok := <-s.outData:
			if !ok {
				continue
			}
			go s.Write(data.Head, DATA, data.Body)

		case c := <-s.cClose:
			log.Debug("session close", c)
			this.handle.Close(s.fd)
			this.Close(s.fd)
			s.Release()
			return
		}
	}
}

func (this *TcpClient) heartbeat(s *Session) {
	for t := range this.ticker.C {
		s.Write(this.MsgCounter.GetNum(), HEARTBEAT, []byte{})

		select {
		case sid := <-this.cHeartbeat:
			log.Debug("heartbeat ret:", t, sid)

		case <-time.After(3 * time.Second):
			log.Debug("Timed out")
		}
	}
}

func (this *TcpClient) handleHeartbeatRet(fd uint16, sessionID uint16) {
	this.cHeartbeat <- sessionID
}

func (this *TcpClient) Close(uint16) {
	this.ticker.Stop()
	close(this.cHeartbeat)
}
