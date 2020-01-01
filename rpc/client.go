package rpc

import (
	"eegos/gate"
	"eegos/log"

	"encoding/json"
	"time"
)

type Client struct {
	callRet   map[uint16](chan []interface{})
	tcpClient *gate.TcpClient
	outData   chan *gate.Data
}

func NewClient() *Client {
	newClient := &Client{}
	newClient.callRet = make(map[uint16](chan []interface{}))
	newClient.tcpClient = gate.NewTcpClient(newClient)
	return newClient
}

func (this *Client) Dial(addr string) {
	this.tcpClient.Dial(addr)
}

func (this *Client) Connect(fd uint16, outData chan *gate.Data) {
	this.outData = outData
}

func (this *Client) Message(fd uint16, sessionID uint16, body []byte) {
	waitRet, ok := this.callRet[sessionID]
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
	delete(this.callRet, sessionID)
	close(waitRet)
}

func (this *Client) Heartbeat(uint16, uint16) {
}

func (this *Client) Close(uint16) {
	for sessionID, waitRet := range this.callRet {
		delete(this.callRet, sessionID)
		close(waitRet)
	}
}

func (this *Client) Call(v []interface{}) []interface{} {
	sessionID := this.tcpClient.MsgCounter.GetNum()
	//TODO make a channel list pool
	waitRet := make(chan []interface{})

	body, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	this.callRet[sessionID] = waitRet

	this.outData <- &gate.Data{Head: sessionID, Body: body}

	select {
	case ret, ok := <-waitRet:
		if ok {
			return ret
		}

	case <-time.After(3 * time.Second):
		log.Debug("Timed out")
	}
	log.Debug("some problems on server")
	return nil
}

func (this *Client) Send(v []interface{}) {
	sessionID := this.tcpClient.MsgCounter.GetNum()
	body, err := json.Marshal(v)
	if err != nil {
		return
	}
	this.outData <- &gate.Data{Head: sessionID, Body: body}
}

func (this *Client) SendCallBack(v []interface{}, callBack func(callRet interface{})) {
}
