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
	this.session.conn.Close()
	close(this.cHeartBeat)
	for session_id, waitRet := range this.callRet {
		delete(this.callRet, session_id)
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
					waitRet, ok := this.callRet[session_id]
					if !ok {
						continue
					}
					waitRet <- data.body
					delete(this.callRet, session_id)
					close(waitRet)
				}
			}

		case <-session.cClose:
			log.Println("session close")
			this.CloseSession()
			break
		}
	}
}

func (this *Client) Call(v []interface{}) []interface{} {
	session_id := counter.GetNum()
	//TODO make a channel list pool
	waitRet := make(chan []interface{})
	this.callRet[session_id] = waitRet
	this.session.HandleWrite(session_id, v)

	select {
	case ret, ok := <-waitRet:
		if ok {
			return ret
		}

	case <-time.After(3 * time.Second):
		log.Println("Timed out")
	}
	log.Println("some problems on server")
	return nil
}

func (this *Client) Send(v []interface{}) {
	session_id := counter.GetNum()
	this.session.HandleWrite(session_id, v)
}
