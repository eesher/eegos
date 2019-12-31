package cluster

import (
	"eegos/rpc"

	"strings"
)

type ServerInfo struct {
	server *rpc.Server
	port   string
}

var cServer *ServerInfo

func Open(addr string) {
	server := rpc.NewServer()
	server.Open(addr)
	indx := strings.LastIndex(addr, ":")
	cServer = &ServerInfo{server: server, port: addr[indx+1:]}
}

func Register(rcvr interface{}) {
	cServer.server.Register(rcvr)
}

/*
var cClient map[string]*rpc.Client

func Connect(serverName string, addr string) {
	if cClient == nil {
		cClient = make(map[string]*rpc.Client)
	}
	if cClient[serverName] == nil {
		client := rpc.NewClient()
		client.Dial(addr)
		cClient[serverName] = client
	}
}

func Call(serverName string, v ...interface{}) []interface{} {
	client := cClient[serverName]
	if client == nil {
		panic("cannot find server:" + serverName)
	}
	return client.Call(v)
}

func Send(serverName string, v ...interface{}) {
	client := cClient[serverName]
	if client == nil {
		panic("cannot find server:" + serverName)
	}
	client.Send(v)
}
*/
