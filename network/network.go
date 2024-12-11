package network

const (
	NEW_CONNECTION = iota
	WORKING
	CLOSING
	CLOSED
)

const (
	PKG_TYPE = iota
	HEARTBEAT
	HEARTBEAT_RET
	DATA
)

type Data struct {
	dType uint8
	head  uint16
	body  []byte
}

type Handler interface {
	Connect(uint16, Session)
	Message(uint16, uint16, []byte)
	Heartbeat(uint16, uint16)
	Close(uint16)
}

type Session interface {
	Forward(func(uint16, uint16, []byte))
	Close()
	doWrite(uint16, uint8, []byte)
}
