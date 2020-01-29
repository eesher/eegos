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
