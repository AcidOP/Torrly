package messages

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const (
	IDChoke = iota
	IDUnchoke
	IDInterested
	IDNotInterested
	IDHave
	IDBitfield
	IDRequest
	IDPiece
	IDCancel
	IDPort
)

type Message struct {
	ID      int
	Payload []byte
}

func MessageIDToString(id int) string {
	switch id {
	case IDChoke:
		return "Choke"
	case IDUnchoke:
		return "Unchoke"
	case IDInterested:
		return "Interested"
	case IDNotInterested:
		return "NotInterested"
	case IDHave:
		return "Have"
	case IDBitfield:
		return "Bitfield"
	case IDRequest:
		return "Request"
	case IDPiece:
		return "Piece"
	case IDCancel:
		return "Cancel"
	case IDPort:
		return "Port"
	default:
		return "Unknown Message ID"
	}
}

func IsValidMessageID(id int) bool {
	switch id {
	case IDChoke, IDUnchoke, IDInterested, IDNotInterested,
		IDHave, IDBitfield, IDRequest, IDPiece, IDCancel, IDPort:
		return true
	default:
		return false
	}
}

func ReceivePeerMessage(conn net.Conn) (*Message, error) {
	lengthBuffer := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBuffer); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuffer)

	if length == 0 { // Keep Alive Message
		return nil, nil
	}

	msgBuf := make([]byte, length-1)
	if _, err := io.ReadFull(conn, msgBuf); err != nil {
		return nil, err
	}

	msgID := int(msgBuf[0])
	payload := msgBuf[1:]

	if msgID == IDBitfield {
		expectedLength := (2680 + 7) / 8
		if len(payload) > expectedLength {
			fmt.Println("Bitfield too long! Possible error.")
		}
	}

	fmt.Println("Raw MessageID: ", msgID)

	fmt.Printf("Received message ID: %s, Length: %d\n", MessageIDToString(msgID), length)
	fmt.Printf("Payload: %x\n", payload)

	return &Message{ID: msgID, Payload: payload}, nil
}
