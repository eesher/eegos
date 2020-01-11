package rpc

import (
	"eegos/gate"
	"eegos/log"

	"encoding/json"
	"reflect"
	"strings"
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
	tcpServer  *gate.TcpServer
	sessions   map[uint16]*gate.Session
}

func NewServer(addr string) *Server {
	services := make(map[string]*Service)
	newServer := Server{serviceMap: services}
	newServer.tcpServer = gate.NewTcpServer(&newServer, addr)
	newServer.sessions = make(map[uint16]*gate.Session)

	return &newServer
}

func (this *Server) Start() {
	this.tcpServer.Start()
}

func (this *Server) Connect(fd uint16, session *gate.Session) {
	log.Debug("rpc server new connection", fd)
	this.sessions[fd] = session
}

func (this *Server) Message(fd uint16, sessionID uint16, body []byte) {
	s, ok := this.sessions[fd]
	if !ok {
		log.Error("fd not found", fd, sessionID)
		return
	}

	args := []interface{}{}
	err := json.Unmarshal(body, &args)
	if err != nil {
		log.Error(err)
		return
	}
	//log.Debug("get message", sessionID, args[0])

	if len(args) == 0 {
		log.Error("no args")
		return
	}

	info := args[0].(string)

	dot := strings.LastIndex(info, ".")
	if dot < 0 {
		log.Error("cannot find dot")
		return
	}
	serviceName := info[:dot]
	methodName := info[dot+1:]

	sInfo := this.serviceMap[serviceName]
	if sInfo == nil {
		log.Error("cannot find service")
		return
	}
	mInfo := sInfo.methods[methodName]
	if mInfo == nil {
		log.Error("cannot find method", methodName, len(methodName))
		return
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
	//log.Debug("run func", sessionID, args[0])
	ret := this.RunFunc(mInfo, callArgs)
	if ret != nil {
		retBody, err := json.Marshal(ret)
		if err != nil {
			log.Error(err)
			return
		}
		this.tcpServer.Write(s, sessionID, retBody)
	} else {
		this.tcpServer.Write(s, sessionID, nil)
	}
	//retPkg := &gate.Data{Head: sessionID, Body: retBody}
	//log.Debug("return", sessionID, args[0])
	//this.tcpGate.Write(s, sessionID, retBody)
	//outc <- retPkg
}

func (this *Server) Heartbeat(fd uint16, sessionID uint16) {
	log.Debug("Heartbeat", fd, sessionID)
}

func (this *Server) Close(fd uint16) {
	log.Debug("need close session", fd)
}

/*
func (this *Server) Open(addr string) {
	//this.tcpGate.Open(addr)
}
*/

func (this *Server) RunFunc(m *methodType, args []reflect.Value) []interface{} {
	callRet := m.method.Func.Call(args)

	retLen := len(callRet)
	if retLen == 0 {
		return nil
	}
	reply := make([]interface{}, retLen)
	for i := 0; i < retLen; i++ {
		reply[i] = callRet[i].Interface()
	}

	return reply
	//session.HandleWrite(sessionID, reply)
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
		log.Debug("registered method: ", mname, len(methodInfo.args))
	}
	this.serviceMap[s.name] = s
	log.Debug("registered Service: ", s.name)
}
