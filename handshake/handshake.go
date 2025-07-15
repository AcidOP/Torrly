package handshake

import (
	"bytes"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	HASH_LENGTH     = 68
	PROTOCOL_LENGTH = 19
	PROTOCOL_STRING = "BitTorrent protocol"
)

// https://wiki.theory.org/BitTorrentSpecification#Handshake
type Handshake struct {
	InfoHash  string
	pId       string
	pLength   int
	pStr      string
	pReserved string
}

func NewHandshake(infoHash, peerId string) *Handshake {
	return &Handshake{
		pLength:   PROTOCOL_LENGTH,
		pStr:      PROTOCOL_STRING,
		pReserved: strings.Repeat("\x00", 8),
		InfoHash:  infoHash,
		pId:       peerId,
	}
}

// Takes a Connection (to another peer) as an argument and sends our handshake.
// Then waits for the peer to respond with its handshake and return it
func (own Handshake) ExchangeHandshake(connPeer net.Conn) ([]byte, error) {
	hBuf := bytes.Buffer{}
	hBuf.WriteByte(byte(own.pLength)) // pstrlen (1 byte)
	hBuf.WriteString(own.pStr)        // pstr (19 bytes)
	hBuf.Write([]byte(own.pReserved)) // reserved (8 bytes)
	hBuf.Write([]byte(own.InfoHash))  // info_hash (20 bytes)
	hBuf.Write([]byte(own.pId))       // peer_id (20 bytes)

	hBytes := hBuf.Bytes() // 1 + 19 + 8 + 20 + 20 = 68 bytes

	if len(hBytes) != HASH_LENGTH {
		return nil, errors.New("handshake byte length is expected 68 bytes, got: " + strconv.Itoa(len(hBytes)))
	}

	// fmt.Printf("\nSending Handshake: %x\n", hBytes)

	if _, err := connPeer.Write(hBytes); err != nil {
		return nil, errors.New("failed to send handshake: " + err.Error())
	}

	connPeer.SetReadDeadline(time.Now().Add(time.Second * 5))

	received := make([]byte, HASH_LENGTH)
	if _, err := io.ReadFull(connPeer, received); err != nil {
		return nil, errors.New("invalid handshake response from peer: " + err.Error())
	}

	// fmt.Printf("\nReceived following handshake %s\n", buf)
	return received, nil
}

// Decode a Handshake sent by another Peer
func DecodeHandshake(buf []byte) (*Handshake, error) {
	if len(buf) != HASH_LENGTH {
		return nil, errors.New("invalid handshake length, expected 68 bytes, got: " + strconv.Itoa(len(buf)))
	}

	h := Handshake{
		pLength:   int(buf[0]),        // 1 byte
		pStr:      string(buf[1:20]),  // "BitTorrent protocol" 19 bytes
		pReserved: string(buf[20:28]), // 8 bytes
		InfoHash:  string(buf[28:48]), // 20 bytes
		pId:       string(buf[48:68]), // 20 bytes
	}

	if h.pLength != PROTOCOL_LENGTH || h.pStr != PROTOCOL_STRING {
		return nil, errors.New("invalid handshake protocol string")
	}

	return &h, nil
}

func (own *Handshake) VerifyHandshake(raw []byte) bool {
	h2, err := DecodeHandshake(raw)
	if err != nil {
		return false
	}

	if h2.pLength != PROTOCOL_LENGTH ||
		h2.pStr != PROTOCOL_STRING ||
		len(h2.pReserved) != 8 ||
		own.InfoHash != h2.InfoHash {
		return false
	}
	return true
}
