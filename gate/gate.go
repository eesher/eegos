package gate

const (
	NEW_CONNECTION = iota
	WORKING
	CLOSING
	CLOSED
)

const (
	HEARTBEAT = iota
	HEARTBEAT_RET
	DATA
)

type Data struct {
	dType uint8
	Head  uint16
	Body  []byte
}
