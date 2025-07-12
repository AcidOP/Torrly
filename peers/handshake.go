package peers

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"
)

const HASH_LENGTH = 68

// https://wiki.theory.org/BitTorrentSpecification#Handshake
type Handshake struct {
	InfoHash  string
	pId       string
	pLength   int
	pStr      string
	pReserved string
}

func (p Peer) HandshakePeer(iHash, peerId string) ([]byte, error) {
	c, err := p.ConnectToPeer()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	h := Handshake{
		pLength:   19, // Decimal 19
		pStr:      "BitTorrent protocol",
		pReserved: "\x00\x00\x00\x00\x00\x00\x00\x00",
		InfoHash:  iHash,
		pId:       peerId,
	}

	hBuf := bytes.Buffer{}
	hBuf.WriteByte(byte(h.pLength)) // pstrlen (1 byte)
	hBuf.WriteString(h.pStr)        // pstr (19 bytes)
	hBuf.Write([]byte(h.pReserved)) // reserved (8 bytes)
	hBuf.Write([]byte(h.InfoHash))  // info_hash (20 bytes)
	hBuf.Write([]byte(h.pId))       // peer_id (20 bytes)

	hBytes := hBuf.Bytes() // 1 + 19 + 8 + 20 + 20 = 68 bytes

	if len(hBytes) != HASH_LENGTH {
		return nil, errors.New("handshake byte length is not 68 bytes, got: " + strconv.Itoa(len(hBytes)))
	}

	// fmt.Printf("\nSending Handshake: %x\n", hBytes)

	if _, err := c.Write(hBytes); err != nil {
		return nil, errors.New("failed to send handshake: " + err.Error())
	}

	// Allow the peer 5 second time to send a hash back.
	if err = c.SetReadDeadline(time.Now().Add(time.Second * 5)); err != nil {
		fmt.Println("failed to set read deadline:", err)
		return nil, err
	}

	// Accept a handshake from peer
	buf := make([]byte, HASH_LENGTH)
	if _, err = c.Read(buf); err != nil || len(buf) != HASH_LENGTH {
		return nil, errors.New("invalid handshake response from peer: " + err.Error())
	}

	// fmt.Printf("\nReceived following handshake %s\n", buf)
	return buf, nil
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

	if h.pLength != 19 || h.pStr != "BitTorrent protocol" {
		return nil, errors.New("invalid handshake protocol string")
	}

	fmt.Println("Length:", h.pLength)
	fmt.Println("Protocol String:", h.pStr)
	fmt.Println("Reserved:", h.pReserved)
	fmt.Printf("Info Hash: %x\n", h.InfoHash)
	fmt.Println("Peer ID:", h.pId)

	fmt.Println()

	return &h, nil
}

func ValidateHandshake(h1, h2 *Handshake) bool {
	if h1.pLength != h2.pLength ||
		h1.pStr != h2.pStr ||
		h1.pReserved != h2.pReserved ||
		h1.InfoHash != h2.InfoHash ||
		h1.pId != h2.pId {
		return false
	}
	return true
}
