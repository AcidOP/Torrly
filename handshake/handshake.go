package handshake

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	HANDSHAKE_LENGTH = 68
	PROTOCOL_LENGTH  = 19
	PROTOCOL_STRING  = "BitTorrent protocol"
	RESERVED_LENGTH  = 8
	HASH_LENGTH      = 20
	PEER_ID_LENGTH   = 20
)

// https://wiki.theory.org/BitTorrentSpecification#Handshake
type Handshake struct {
	pLength   int
	pStr      string
	pReserved []byte
	InfoHash  []byte
	PeerID    []byte
}

func NewHandshake(infoHash, peerID []byte) (*Handshake, error) {
	if len(infoHash) != HASH_LENGTH {
		return nil, fmt.Errorf("info hash must be %d bytes, got %d", HASH_LENGTH, len(infoHash))
	}

	if len(peerID) != PEER_ID_LENGTH {
		return nil, fmt.Errorf("peer id must be %d bytes, got %d", PEER_ID_LENGTH, len(peerID))
	}

	return &Handshake{
		pLength:   PROTOCOL_LENGTH,
		pStr:      PROTOCOL_STRING,
		pReserved: bytes.Repeat([]byte{0x00}, RESERVED_LENGTH),
		InfoHash:  infoHash,
		PeerID:    peerID,
	}, nil
}

func (h *Handshake) Serialize() []byte {
	hBuf := bytes.Buffer{}

	// Build handshake: pstrlen + pstr + reserved + info_hash + peer_id
	hBuf.WriteByte(byte(h.pLength)) // pstrlen (1 byte)
	hBuf.WriteString(h.pStr)        // pstr (19 bytes)
	hBuf.Write(h.pReserved)         // reserved (8 bytes)
	hBuf.Write(h.InfoHash)          // info_hash (20 bytes)
	hBuf.Write(h.PeerID)            // peer_id (20 bytes)

	return hBuf.Bytes() // 1 + 19 + 8 + 20 + 20 = 68 bytes
}

func (h *Handshake) String() string {
	return string(h.Serialize())
}

// Takes a Connection (to another peer) as an argument and sends our handshake.
// Then waits for the peer to respond with its handshake and return it
func (h *Handshake) ExchangeHandshake(conn net.Conn) error {
	if _, err := conn.Write(h.Serialize()); err != nil {
		return fmt.Errorf("failed to send handshake: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	defer conn.SetReadDeadline(time.Time{})

	received := make([]byte, HANDSHAKE_LENGTH)

	if _, err := io.ReadFull(conn, received); err != nil {
		return fmt.Errorf("failed to read handshake response: %v", err)
	}

	hs, err := DecodeHandshake(received)
	if err != nil {
		return fmt.Errorf("failed to decode handshake: %v", err)
	}

	return h.VerifyHandshake(hs)
}

// Decode a Handshake sent by another Peer
func DecodeHandshake(buf []byte) (*Handshake, error) {
	if len(buf) != HANDSHAKE_LENGTH {
		return nil, fmt.Errorf("invalid handshake length, expected %d bytes, got: %d",
			HANDSHAKE_LENGTH, len(buf))
	}

	protoLength := int(buf[0])
	if protoLength != PROTOCOL_LENGTH {
		return nil, fmt.Errorf("invalid protocol length: expected %d, got %d",
			PROTOCOL_LENGTH, protoLength)
	}

	protoStr := string(buf[1 : 1+protoLength])
	if protoStr != PROTOCOL_STRING {
		return nil, fmt.Errorf("invalid protocol string: expected %q, got %q",
			PROTOCOL_STRING, protoStr)
	}

	h := &Handshake{
		pLength:   protoLength,
		pStr:      protoStr,
		pReserved: buf[20:28],
		InfoHash:  buf[28:48],
		PeerID:    buf[48:68],
	}

	return h, nil
}

func (h *Handshake) VerifyHandshake(h2 *Handshake) error {
	if h2.pLength != PROTOCOL_LENGTH {
		return fmt.Errorf("protocol length mismatch: expected %d, got %d",
			PROTOCOL_LENGTH, h2.pLength)
	}

	if h2.pStr != PROTOCOL_STRING {
		return fmt.Errorf("protocol string mismatch: expected %q, got %q",
			PROTOCOL_STRING, h2.pStr)
	}

	if len(h2.pReserved) != RESERVED_LENGTH {
		return fmt.Errorf("reserved field length mismatch: expected %d, got %d",
			RESERVED_LENGTH, len(h2.pReserved))
	}

	if !bytes.Equal(h.InfoHash, h2.InfoHash) {
		return fmt.Errorf("info hash mismatch: expected %x, got %x",
			h.InfoHash, h2.InfoHash)
	}

	return nil
}
