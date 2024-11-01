package websocket

import "io"

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
