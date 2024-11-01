package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"io"

	"github.com/radenrishwan/aci/chttp"
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
	key := ""

	request, err := chttp.NewRequest(conn)
	if err != nil {
		return client, err
	}

	if _, ok := request.Headers["sec-websocket-key"]; ok {
		key = request.Headers["sec-websocket-key"]
	}

	if _, ok := request.Headers["Sec-WebSocket-Key"]; ok {
		key = request.Headers["Sec-WebSocket-Key"]
	}

	if key == "" {
		return client, cwebsocket.NewWsError("Sec-WebSocket-Key is required", "")
	}

	acceptKey := generateWebsocketKey(key)

	_, err = conn.Write([]byte(
		"HTTP/1.1 101 Switching Protocols\r\n" +
			"Upgrade: websocket\r\n" +
			"Connection: Upgrade\r\n" +
			"Sec-WebSocket-Accept: " + acceptKey + "\r\n" +
			"\r\n",
	))

	if err != nil {
		return client, cwebsocket.NewWsError("Error while upgrading connection : ", err.Error())
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

func generateWebsocketKey(key string) string {
	sha := sha1.New()
	sha.Write([]byte(key))
	sha.Write([]byte(MAGIC_KEY))

	return base64.StdEncoding.EncodeToString(sha.Sum(nil))
}
