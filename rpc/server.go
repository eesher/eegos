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
	tcpGate    *gate.TcpServer
	outDatas   map[uint16]chan *gate.Data
}

func NewServer() *Server {
	services := make(map[string]*Service)
	newServer := Server{serviceMap: services}
	newServer.tcpGate = gate.NewTcpServer(&newServer)
	newServer.outDatas = make(map[uint16]chan *gate.Data)

	return &newServer
}

func (this *Server) Connect(fd uint16, outData chan *gate.Data) {
	log.Debug("rpc server new connection", fd)
	this.outDatas[fd] = outData
}

func (this *Server) Message(fd uint16, sessionID uint16, body []byte) {
	outc, ok := this.outDatas[fd]
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
	ret := this.RunFunc(mInfo, callArgs)
	retBody, err := json.Marshal(ret)
	if err != nil {
		log.Error(err)
		return
	}
	retPkg := &gate.Data{Head: sessionID, Body: retBody}
	outc <- retPkg
}

func (this *Server) HeartBeat(fd uint16, sessionID uint16) {
	log.Debug("HeartBeat", fd, sessionID)
}

func (this *Server) Close(fd uint16) {
	log.Debug("rpc server close", fd)
}

func (this *Server) Open(addr string) {
	this.tcpGate.Open(addr)
}

func (this *Server) RunFunc(m *methodType, args []reflect.Value) []interface{} {
	callRet := m.method.Func.Call(args)

	retLen := len(callRet)
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
