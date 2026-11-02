package peers

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/AcidOP/torrly/messages"
)

// Read function reads a `messages.Message` from the peer's connection.
// (Optionally) accepts a timeout duration to set a read deadline.
// If no timeout is provided, it defaults to 2 minutes.
func (p *Peer) Read(timeout ...time.Duration) (*messages.Message, error) {
	duration := 120 * time.Second
	if len(timeout) > 0 {
		duration = timeout[0]
	}

	p.conn.SetReadDeadline(time.Now().Add(duration))
	defer p.conn.SetReadDeadline(time.Time{})

	return messages.Receive(p.conn)
}

func (p *Peer) send(msg *messages.Message) error {
	if _, err := p.conn.Write(msg.Serialize()); err != nil {
		return err
	}
	fmt.Printf("\nSent message (%s) to peer: %s\n", msg, p.IP)
	return nil
}

func (p *Peer) SendInterested() error {
	msg := messages.Message{ID: messages.MsgInterested}
	return p.send(&msg)
}

func (p *Peer) SendNotInterested() error {
	msg := messages.Message{ID: messages.MsgNotInterested}
	return p.send(&msg)
}

func (p *Peer) SendCancel() error {
	msg := messages.Message{ID: messages.MsgCancel}
	return p.send(&msg)
}

func (p *Peer) SendKeepAlive() error {
	msg := messages.Message{ID: messages.MsgKeepAlive}
	return p.send(&msg)
}

// sendRequest sends a request message to the peer for a specific piece.
// index: The index of the piece to request.
// length: The length (normally 16 KB) of the piece to request.
// begin: The offset within the piece to start the request.
// https://wiki.theory.org/BitTorrentSpecification#Messages
func (p *Peer) SendRequest(index, length, begin int) error {
	msg := messages.Message{
		ID:      messages.MsgRequest,
		Payload: make([]byte, 12),
	}
	binary.BigEndian.PutUint32(msg.Payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(msg.Payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(msg.Payload[8:12], uint32(length))

	return p.send(&msg)
}

// bytesToBoolSlice helper func converts a []byte bitfield to a []bool slice.
func bytesToBoolSlice(bf []byte) []bool {
	bools := make([]bool, 0, len(bf)*8)
	for _, b := range bf {
		for i := 7; i >= 0; i-- {
			bools = append(bools, (b>>i)&1 == 1)
		}
	}
	return bools
}
