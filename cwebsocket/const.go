package cwebsocket

import "fmt"

const MAGIC_KEY = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type MessageType int

const (
	TEXT MessageType = iota
	BINARY
	PING
	PONG
	CLOSE
)

const (
	MESSAGE_FRAME = 0x81
	BINARY_FRAME  = 0x82
	PING_FRAME    = 0x89
	PONG_FRAME    = 0x8A
	CLOSE_FRAME   = 0x88
)

const (
	STATUS_CLOSE_NORMAL_CLOSURE        = 1000
	STATUS_CLOSE_GOING_AWAY            = 1001
	STATUS_CLOSE_PROTOCOL_ERR          = 1002
	STATUS_CLOSE_UNSUPPORTED           = 1003
	STATUS_CLOSE_NO_STATUS             = 1005
	STATUS_CLOSE_ABNORMAL_CLOSURE      = 1006
	STATUS_CLOSE_INVALID_PAYLOAD       = 1007
	STATUS_CLOSE_POLICY_VIOLATION      = 1008
	STATUS_CLOSE_MESSAGE_TOO_BIG       = 1009
	STATUS_CLOSE_MANDATORY_EXTENSION   = 1010
	STATUS_CLOSE_INTERNAL_SERVER_ERROR = 1011
	STATUS_CLOSE_SERVICE_RESTART       = 1012
	STATUS_CLOSE_TRY_AGAIN_LATER       = 1013
	STATUS_CLOSE_TLS_HANDSHAKE         = 1015
)

type WsError struct {
	Msg    string
	Reason string
}

func (e *WsError) Error() string {
	return fmt.Sprintf("%s: %s", e.Msg, e.Reason)
}

func NewWsError(msg string, reason string) *WsError {
	return &WsError{Msg: msg, Reason: reason}
}

type WsFrame struct {
	Fin     bool
	RSV1    bool
	RSV2    bool
	RSV3    bool
	Opcode  uint8
	Mask    bool
	Length  uint64
	MaskKey [4]byte
	Payload []byte
}
