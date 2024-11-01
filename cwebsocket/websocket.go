package cwebsocket

import (
	"encoding/binary"
	"io"
)

// write a websocket frame to the connection
func Write(conn io.Writer, msg []byte) error {
	frame := EncodeFrame(msg, TEXT)

	_, err := conn.Write(frame)
	if err != nil {
		return NewWsError("Error sending message : ", err.Error())
	}

	return nil
}

// write a string message to the connection
func WriteString(conn io.Writer, msg string) error {
	return Write(conn, []byte(msg))
}

// write a websocket frame to the connection with a specific message type
func WriteWithMessageType(conn io.Writer, msg string, messageType MessageType) error {
	frame := EncodeFrame([]byte(msg), messageType)

	_, err := conn.Write(frame)
	if err != nil {
		return NewWsError("Error sending message : ", err.Error())
	}

	return nil
}

// encode a websocket frame to be sent over the connection
func EncodeFrame(msg []byte, messageType MessageType) []byte {
	frame := make([]byte, 0)
	switch messageType {
	case TEXT:
		frame = append(frame, MESSAGE_FRAME)
	case BINARY:
		frame = append(frame, BINARY_FRAME)
	case PING:
		frame = append(frame, PING_FRAME)
	case PONG:
		frame = append(frame, PONG_FRAME)
	case CLOSE:
		frame = append(frame, CLOSE_FRAME)
	default:
		frame = append(frame, MESSAGE_FRAME)
	}

	length := len(msg)
	if length < 126 {
		frame = append(frame, byte(length))
	} else if length <= 0xFFFF {
		frame = append(frame, 126)

		// add length as 16-bit unsigned integer
		frame = append(frame, byte(length>>8))
		frame = append(frame, byte(length&0xFF))
	} else {
		frame = append(frame, 127)

		// add length as 64-bit unsigned integer
		for i := 7; i >= 0; i-- {
			frame = append(frame, byte(length>>(i*8)))
		}
	}

	frame = append(frame, msg...)
	return frame
}

// decode a websocket frame from the connection
func DecodeFrame(data []byte) (*WsFrame, error) {
	if len(data) < 2 {
		return nil, NewWsError("insufficient data for frame", "")
	}

	frame := &WsFrame{}
	frame.Fin = (data[0] & 0x80) != 0
	frame.RSV1 = (data[0] & 0x40) != 0
	frame.RSV2 = (data[0] & 0x20) != 0
	frame.RSV3 = (data[0] & 0x10) != 0
	frame.Opcode = data[0] & 0x0F

	frame.Mask = (data[1] & 0x80) != 0
	payloadLength := uint64(data[1] & 0x7F)

	var dataOffset uint64
	switch payloadLength {
	case 126:
		if len(data) < 4 {
			return nil, NewWsError("insufficient data for payload length", "")
		}
		frame.Length = uint64(binary.BigEndian.Uint16(data[2:4]))
		dataOffset = 4
	case 127:
		if len(data) < 10 {
			return nil, NewWsError("insufficient data for payload length", "")
		}
		frame.Length = binary.BigEndian.Uint64(data[2:10])
		dataOffset = 10
	default:
		frame.Length = payloadLength
		dataOffset = 2
	}

	if frame.Mask {
		if uint64(len(data)) < dataOffset+4 {
			return nil, NewWsError("insufficient data for mask key", "")
		}
		copy(frame.MaskKey[:], data[dataOffset:dataOffset+4])
		dataOffset += 4
	}

	if uint64(len(data)) < dataOffset+frame.Length {
		return nil, NewWsError("insufficient data for payload", "")
	}
	payload := data[dataOffset : dataOffset+frame.Length]

	if frame.Mask {
		unmaskedPayload := make([]byte, len(payload))
		for i, b := range payload {
			unmaskedPayload[i] = b ^ frame.MaskKey[i%4]
		}
		payload = unmaskedPayload
	}

	frame.Payload = payload

	return frame, nil
}

// close the connection with a specific reason and status code
func Close(conn io.WriteCloser, reason string, code int) error {
	// send close normal closue
	closeMSG := make([]byte, 0)

	// add status code on the first 2 byte
	closeMSG = append(closeMSG, byte(code>>8))
	closeMSG = append(closeMSG, byte(code&0xFF))

	// add reason
	closeMSG = append(closeMSG, []byte(reason)...)

	frame := EncodeFrame(closeMSG, CLOSE)

	_, err := conn.Write(frame)
	if err != nil {
		return NewWsError("Error sending close signal : ", err.Error())
	}

	err = conn.Close()
	if err != nil {
		return NewWsError("Error closing connection : ", err.Error())
	}

	return nil
}
