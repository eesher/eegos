package rpc

import (
	"log"
	"net"
	"sync"
)

type Counter struct {
	sync.Mutex
	num uint
}

var counter = Counter{num: 0}

type Client struct {
	session *Session
	callRet map[uint](chan []interface{})
}

func (this *Counter) GetNum() uint {
	var num uint
	this.Lock()
	if this.num >= 65535 {
		this.num = 0
	} else {
		this.num++
	}
	num = this.num
	this.Unlock()
	return num
}

func NewClient() *Client {
	client := new(Client)
	client.callRet = make(map[uint](chan []interface{}))
	return client
}

func (this *Client) Dial(addr string) {
	//tcpAddr, _ := net.ResolveTCPAddr("tcp4", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println("cluster.Call: net.Dial: ", err.Error())
		return
	}
	this.session = CreateSession(conn)
	go this.HandleData(this.session)
}

func (this *Client) HandleData(session *Session) {
	for {
		select {
		case data, ok := <-session.newData:
			if !ok {
				log.Println("read error:")
			} else {
				session_id := data.head
				waitRet := this.callRet[session_id]
				if waitRet != nil {
					waitRet <- data.body
				}
			}

		case <-session.cClose:
			log.Println("session close")
			break
		}

		//heartbeat
	}
}

func (this *Client) Call(v []interface{}) (ret []interface{}) {
	session_id := counter.GetNum()
	waitRet := make(chan []interface{})
	this.callRet[session_id] = waitRet
	this.session.HandleWrite(session_id, v)

	select {
	case ret = <-waitRet:
		close(waitRet)
		this.callRet[session_id] = nil
		return
	}
}

func (this *Client) Send(v []interface{}) {
	session_id := counter.GetNum()
	this.session.HandleWrite(session_id, v)
}
