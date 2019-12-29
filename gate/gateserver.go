package gate

import (
	"eegos/log"

	"net"
	"strings"
)

type Handler interface {
	Connect(s *Session)
	Message(data *Data)
	Close(s *Session)
}

type TcpServer struct {
	isOpen bool
	handle *Handler
}

func NewTcpServer(handle *Handler) *TcpServer {
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

func (this *TcpServer) NewSession(conn net.TCPConn) {
	log.Debug("new connection from ", conn.RemoteAddr())
	session := CreateSession(conn)
	//handle new connection
	this.handle.Connect(session)
}
