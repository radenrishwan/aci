package websocket

import "io"

const MAGIC_KEY = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type Websocket struct {
	Option *WSOption
}

type Client struct {
	Conn   io.ReadWriteCloser
	option *WSOption
}

type WSOption struct {
	MsgMaxSize int
}

var DefaultWSOption = WSOption{
	MsgMaxSize: 1024,
}
