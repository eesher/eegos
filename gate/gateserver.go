package gate

import (
	"eegos/log"

	"net"
	"time"
)

type Handler interface {
	Connect(uint16, chan *Data)
	Message(uint16, uint16, []byte)
	HeartBeat(uint16, uint16)
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
				go this.handle.HeartBeat(s.fd, data.Head)
			case HEARTBEAT_RET:
				go this.handle.HeartBeat(s.fd, data.Head)
			default:
				go this.handle.Message(s.fd, data.Head, data.Body)
			}
		case data, ok := <-s.outData:
			if !ok {
				continue
			}
			go s.Write(data.Head, DATA, data.Body)

		case <-s.cClose:
			log.Debug("session close")
			this.handle.Close(s.fd)
			s.Release()
			break
		}
	}
}
