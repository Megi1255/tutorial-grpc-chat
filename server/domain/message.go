package domain

type Message struct {
	Type MessageType
	Name string
	Text string
}

type MessageType int

const (
	MSGTYPE_UNKNOWN = iota
	MSGTYPE_LOGIN
	MSGTYPE_LOGOUT
	MSGTYPE_MESSAGE
	MSGTYPE_SHUTDOWN
)
