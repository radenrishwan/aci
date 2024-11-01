package websocket

import (
	"io"

	"github.com/radenrishwan/aci/cwebsocket"
)

func NewWebsocket(option *WSOption) (ws Websocket) {
	if option == nil {
		option = &DefaultWSOption
	}

	return Websocket{
		Option: option,
	}
}

func (ws *Websocket) Upgrade(conn io.ReadWriteCloser) (client Client, err error) {
	err = cwebsocket.Upgrade(conn)
	if err != nil {
		return Client{}, err
	}

	client.Conn = conn
	client.option = ws.Option

	return client, nil
}

func (client *Client) Send(msg string) error {
	return cwebsocket.WriteString(client.Conn, msg)
}

func (client *Client) SendBytes(msg []byte) error {
	return cwebsocket.Write(client.Conn, msg)
}

func (client *Client) SendWithMessageType(msg string, messageType cwebsocket.MessageType) error {
	return cwebsocket.WriteWithMessageType(client.Conn, msg, messageType)
}

func (client *Client) Read() ([]byte, error) {
	buf := make([]byte, client.option.MsgMaxSize)

	n, err := client.Conn.Read(buf)
	if err != nil {
		return nil, cwebsocket.NewWsError("Error reading message : ", err.Error())
	}

	f, err := cwebsocket.DecodeFrame(buf[:n])
	if err != nil {
		return nil, cwebsocket.NewWsError("Error decoding frame : ", err.Error())
	}

	// check if close signal
	if f.Opcode == 0x8 {
		return nil, cwebsocket.NewWsError("Close signal received", "")
	}

	return f.Payload, nil
}

func (client *Client) Close(reason string, code int) error {
	return cwebsocket.Close(client.Conn, reason, code)
}
