package peers

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"
)

const hashLen = 68

// https://wiki.theory.org/BitTorrentSpecification#Handshake
type Handshake struct {
	InfoHash  string
	pId       string
	pLength   int
	pStr      string
	pReserved string
}

func HandshakePeer(p Peer, iHash, peerId string) error {
	c, err := p.ConnPeer()
	if err != nil {
		return err
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

	if len(hBytes) != hashLen {
		return errors.New("handshake byte length is not 68 bytes, got: " + strconv.Itoa(len(hBytes)))
	}

	// fmt.Printf("\nSending Handshake: %x\n", hBytes)

	if _, err := c.Write(hBytes); err != nil {
		return errors.New("failed to send handshake: " + err.Error())
	}

	// Allow the peer 5 second time to send a hash back.
	if err = c.SetReadDeadline(time.Now().Add(time.Second * 5)); err != nil {
		fmt.Println("failed to set read deadline:", err)
		return err
	}

	// Accept a handshake from peer
	buf := make([]byte, hashLen)
	if _, err = c.Read(buf); err != nil || len(buf) != hashLen {
		return errors.New("invalid handshake response from peer: " + err.Error())
	}

	fmt.Printf("\nReceived following handshake %x\n", buf)
	fmt.Printf("\nReceived following handshake %s\n", buf)
	return nil
}
