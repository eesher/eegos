package rpc

import (
	"github.com/eesher/eegos/log"
	"github.com/eesher/eegos/network"

	"encoding/json"
	"sync"
	"time"
)

type Client struct {
	callRet   map[uint16](chan []interface{})
	mapLocker *sync.RWMutex
	tcpClient *network.TcpClient
	session   *network.Session
}

func NewClient() *Client {
	newClient := &Client{}
	newClient.callRet = make(map[uint16](chan []interface{}))
	newClient.mapLocker = &sync.RWMutex{}
	newClient.tcpClient = network.NewTcpClient(newClient)
	return newClient
}

func (this *Client) Dial(addr string) {
	this.tcpClient.Dial(addr)
}

func (this *Client) Connect(fd uint16, s *network.Session) {
	this.session = s
}

func (this *Client) Message(fd uint16, sessionID uint16, body []byte) {
	this.mapLocker.RLock()
	waitRet, ok := this.callRet[sessionID]
	this.mapLocker.RUnlock()
	if !ok {
		return
	}

	args := []interface{}{}
	err := json.Unmarshal(body, &args)
	if err != nil {
		log.Error(err)
		return
	}

	waitRet <- args
	this.mapLocker.Lock()
	delete(this.callRet, sessionID)
	this.mapLocker.Unlock()
	close(waitRet)
}

func (this *Client) Heartbeat(uint16, uint16) {
}

func (this *Client) Close(uint16) {
	this.mapLocker.Lock()
	for sessionID, waitRet := range this.callRet {
		delete(this.callRet, sessionID)
		close(waitRet)
	}
	this.mapLocker.Unlock()
	//this.tcpClient.Close()
}

func (this *Client) Call(v []interface{}) []interface{} {
	sessionID := this.tcpClient.GetSessionID()
	//TODO make a channel list pool
	waitRet := make(chan []interface{})

	body, err := json.Marshal(v)
	if err != nil {
		return nil
	}

	//log.Debug("call", sessionID)
	this.mapLocker.Lock()
	this.callRet[sessionID] = waitRet
	this.mapLocker.Unlock()

	//this.outData <- &network.Data{Head: sessionID, Body: body}
	//sessionID := this.tcpClient.WriteData(this.session, body)
	this.tcpClient.Write(this.session, sessionID, body)

	select {
	case ret, ok := <-waitRet:
		if ok {
			return ret
		}

	case <-time.After(3 * time.Second):
		log.Debug("Timed out", sessionID)
	}
	log.Debug("some problems on server")
	return nil
}

func (this *Client) Send(v []interface{}) {
	//sessionID := this.tcpClient.GetSessionID()
	body, err := json.Marshal(v)
	if err != nil {
		return
	}
	//this.outData <- &network.Data{Head: sessionID, Body: body}
	this.tcpClient.WriteData(this.session, body)
	//log.Debug("send", sessionID)
}

func (this *Client) SendCallBack(v []interface{}, callBack func(callRet interface{})) {
}
