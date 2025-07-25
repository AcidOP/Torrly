package messages

import (
	"encoding/binary"
	"fmt"
	"io"
)

type MsgID = uint8

const (
	MsgChoke MsgID = iota
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
	ID      MsgID
	Payload []byte
}

func ReceiveBitField(r io.Reader) ([]byte, error) {
	msg, err := ReceivePeerMessage(r)
	if err != nil {
		return nil, err
	}

	if msg.ID != MsgBitfield {
		return nil, fmt.Errorf("expected bitfield but got ID %d", msg.ID)
	}

	// fmt.Printf("\nReceived Bitfield: %b\n\n", msg.Payload)

	return msg.Payload, nil
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

	msgID := MsgID(messageBuf[0])
	payload := messageBuf[1:]

	fmt.Printf("Received message ID: %d\n", msgID)

	return &Message{
		ID:      msgID,
		Payload: payload,
	}, nil
}
