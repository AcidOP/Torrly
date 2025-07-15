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
	ID      byte
	Payload []byte
}

func (m Message) String() string {
	return string(m.ID) + string(m.Payload[:])
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

	// idBuffer := make([]byte, 1)
	// if _, err := io.ReadFull(conn, idBuffer); err != nil {
	// 	return nil, err
	// }
	// messageID := idBuffer[0]

	// payload := make([]byte, length-1)
	// if _, err := io.ReadFull(conn, payload); err != nil {
	// 	return nil, err
	// }

	msgBuf := make([]byte, length-1)
	if _, err := io.ReadFull(conn, msgBuf); err != nil {
		return nil, err
	}

	// Step 3: Extract message ID and payload
	msgID := msgBuf[0]
	payload := msgBuf[1:]

	fmt.Println("Raw MessageID: ", msgID)

	fmt.Printf("Received message ID: %s, Length: %d\n", MessageIDToString(int(msgID)), length)
	fmt.Printf("Payload: %x\n", payload)

	return &Message{ID: msgID, Payload: payload}, nil
}
