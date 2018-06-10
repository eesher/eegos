package rpc

import (
	"log"
	"net"
	"reflect"
	"strings"
	//"time"
)

type methodType struct {
	method reflect.Method
	args   []reflect.Type
}

type Service struct {
	name    string                 // name of service
	rcvr    reflect.Value          // receiver of methods for the service
	typ     reflect.Type           // type of the receiver
	methods map[string]*methodType // registered methods
}

type Server struct {
	serviceMap map[string]*Service
	//sessions   map[string]*Session
	//session    *Session
	closed bool
}

func NewServer() *Server {
	services := make(map[string]*Service)
	//sessions := make(map[string]*Session)
	return &Server{serviceMap: services, closed: false}
}

func (this *Server) Open(addr string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		log.Println("cluster.Open: net.ResolveTCPAddr: ", err.Error())
		return
	}
	lis, err := net.ListenTCP("tcp", tcpAddr)

	log.Println("cluster.Open: listening")
	if err != nil {
		log.Println("cluster.Open: net.ListenTCP: ", err.Error())
		return
	}

	go func() {
		for !this.closed {
			conn, err := lis.Accept()
			if err != nil {
				log.Println("cluster.Open: lis.Accept: ", err.Error())
				return
			}
			//conn.SetKeepAlive(true)
			//conn.SetKeepAlivePeriod(5 * time.Second)
			go this.NewSession(conn)
		}
	}()
}

func (this *Server) NewSession(conn net.Conn) {
	log.Println("new connection")
	session := CreateSession(conn)
	//this.session = session
	//this.sessions[conn.RemoteAddr().String()] = session
	for {
		select {
		case data, ok := <-session.newData:
			if !ok {
			} else {
				args := data.body
				session_id := data.head
				if len(args) == 0 {
					continue
				}

				info := args[0].(string)

				dot := strings.LastIndex(info, ".")
				if dot < 0 {
					log.Println("cannot find dot")
					continue
				}
				serviceName := info[:dot]
				methodName := info[dot+1:]

				sInfo := this.serviceMap[serviceName]
				if sInfo == nil {
					log.Println("cannot find service")
					continue
				}
				mInfo := sInfo.methods[methodName]
				if mInfo == nil {
					log.Println("cannot find method", methodName, len(methodName))
					continue
				}

				callArgs := make([]reflect.Value, len(mInfo.args)+1)
				callArgs[0] = sInfo.rcvr

				for i := 0; i < len(mInfo.args); i++ {
					/*
						if reflect.TypeOf(args[i+1]).Kind() != mInfo.args[i].Kind() {
							callArgs[i+1] = reflect.ValueOf(args[i+1]).Convert(mInfo.args[i])
							log.Println("arg type convert")
						} else {
							callArgs[i+1] = reflect.ValueOf(args[i+1])
						}
					*/
					callArgs[i+1] = reflect.ValueOf(args[i+1]).Convert(mInfo.args[i])
				}
				go this.RunFunc(mInfo, callArgs, session, session_id)
			}
		case <-session.cClose:
			log.Println("session close")
			break
		}
	}
}

func (this *Server) RunFunc(m *methodType, args []reflect.Value, session *Session, session_id uint) {
	callRet := m.method.Func.Call(args)

	retLen := len(callRet)
	reply := make([]interface{}, retLen)
	for i := 0; i < retLen; i++ {
		reply[i] = callRet[i].Interface()
	}
	session.HandleWrite(session_id, reply)
}

func (this *Server) Register(rcvr interface{}) {
	if this.serviceMap == nil {
		this.serviceMap = make(map[string]*Service)
	}
	s := new(Service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()

	s.methods = make(map[string]*methodType)
	for m := 0; m < s.typ.NumMethod(); m++ {
		method := s.typ.Method(m)

		mtype := method.Type
		mname := method.Name

		methodInfo := methodType{method: method}
		argNum := mtype.NumIn() - 1
		if argNum > 0 {
			methodInfo.args = make([]reflect.Type, argNum)
			//log.Println("arg num", argNum)

			for a := 0; a < argNum; a++ {
				methodInfo.args[a] = mtype.In(a + 1)
				//log.Println(methodInfo.args[a].Kind())
			}
		}
		s.methods[mname] = &methodInfo
		log.Println("registered method: ", mname, len(methodInfo.args))
	}
	this.serviceMap[s.name] = s
	log.Println("registered Service: ", s.name)
}
