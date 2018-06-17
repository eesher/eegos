package rpc

import (
	"log"
	"net"
	"sync"
	"time"
)

type Counter struct {
	sync.Mutex
	num uint16
}

var counter = Counter{num: 0}

type Client struct {
	session    *Session
	callRet    map[uint16](chan []interface{})
	ticker     *time.Ticker
	cHeartBeat chan bool
}

func (this *Counter) GetNum() uint16 {
	var num uint16
	this.Lock()
	this.num++
	num = this.num
	this.Unlock()
	return num
}

func NewClient() *Client {
	client := new(Client)
	client.callRet = make(map[uint16](chan []interface{}))
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
	ticker := time.NewTicker(5 * time.Second)
	this.cHeartBeat = make(chan bool)
	go func() {
		for t := range ticker.C {
			session_id := counter.GetNum()
			this.session.DoWrite(session_id, heartBeat, nil)

			select {
			case <-this.cHeartBeat:
				log.Println("heartbeat ret:", t)
				delete(this.callRet, session_id)

			case <-time.After(3 * time.Second):
				log.Println("Timed out")
			}
		}
	}()
	this.ticker = ticker
}

func (this *Client) CloseSession() {
	this.ticker.Stop()
	close(this.cHeartBeat)
	for _, waitRet := range this.callRet {
		close(waitRet)
	}
}

func (this *Client) HandleData(session *Session) {
	for {
		select {
		case data, ok := <-session.newData:
			if !ok {
				log.Println("read error:")
			} else {
				session_id := data.head
				if data.data_type == heartBeatRet {
					this.cHeartBeat <- true
				} else {
					waitRet := this.callRet[session_id]
					if waitRet != nil {
						waitRet <- data.body
					}
				}
			}

		case <-session.cClose:
			log.Println("session close")
			this.CloseSession()
			break
		}
	}
}

func (this *Client) Call(v []interface{}) (ret []interface{}) {
	session_id := counter.GetNum()
	waitRet := make(chan []interface{})
	this.callRet[session_id] = waitRet
	this.session.HandleWriteTest(session_id, v)

	select {
	case ret = <-waitRet:
		close(waitRet)
		delete(this.callRet, session_id)
		return

	case <-time.After(3 * time.Second):
		log.Println("Timed out")
	}
	return
}

func (this *Client) Send(v []interface{}) {
	session_id := counter.GetNum()
	this.session.HandleWrite(session_id, v)
}
