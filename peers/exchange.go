package peers

import (
	"encoding/binary"
	"fmt"

	"github.com/AcidOP/torrly/messages"
)

func (p *Peer) send(msg *messages.Message) error {
	_, err := p.conn.Write(msg.Serialize())

	fmt.Printf("\nSent message (%s) to peer: %s\n", msg.String(), p.IP.String())
	return err
}

func (p *Peer) SendInterested() error {
	msg := messages.Message{ID: messages.MsgInterested}
	if err := p.send(&msg); err != nil {
		return fmt.Errorf("failed to send interested message: %w", err)
	}

	fmt.Println("Sent `interested` message to peer:", p.IP.String())
	return nil
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

	if err := p.send(&msg); err != nil {
		fmt.Println("Error writing request:", err)
		return fmt.Errorf("failed to send request message: %w", err)
	}
	return nil
}
