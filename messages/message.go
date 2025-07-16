package messages

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	MsgChoke = iota
	MsgUnchoke
	MsgInterested
	MsgNotInterested
	MsgHave
	MsgBitfield
	MsgRequest
	MsgPiece
	MsgCancel
)

type Message struct {
	ID      int
	Payload []byte
}

func ReceivePeerMessage(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	// keep-alive message
	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	if len(messageBuf) < 1 {
		return nil, fmt.Errorf("invalid message format: message too short")
	}

	msgID := int(messageBuf[0])
	payload := messageBuf[1:]

	fmt.Printf("Received message ID: %d\n", msgID)
	if msgID == MsgBitfield {
		fmt.Printf("Bitfield payload length: %d\n", len(payload))
		fmt.Printf("Bitfield payload (hex): %x\n", payload)
	}

	return &Message{
		ID:      msgID,
		Payload: payload,
	}, nil
}
