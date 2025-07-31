package peers

import (
	"encoding/binary"
	"fmt"

	"github.com/AcidOP/torrly/messages"
)

func (p *Peer) receiveBitField() ([]byte, error) {
	msg, err := messages.Receive(p.conn)
	if err != nil {
		return nil, err
	}

	if msg.ID != messages.MsgBitfield {
		return nil, fmt.Errorf("expected bitfield message (ID 5), but got ID %d", msg.ID)
	}

	if len(msg.Payload) == 0 {
		return nil, fmt.Errorf("received empty bitfield payload from peer %s", p.IP.String())
	}

	// fmt.Printf("\nReceived Bitfield: %x\n\n", msg.Payload)

	return msg.Payload, nil
}

func (p *Peer) sendInterested() error {
	msg := messages.Message{ID: messages.MsgInterested}
	_, err := p.conn.Write(msg.Serialize())

	fmt.Println("Sent `ineterested` message to peer:", p.IP.String())
	return err
}

func (p *Peer) sendRequest(index, length, begin int) {
	msg := messages.Message{
		ID:      messages.MsgRequest,
		Payload: make([]byte, 12),
	}
	binary.BigEndian.PutUint32(msg.Payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(msg.Payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(msg.Payload[8:12], uint32(length))

	_, err := p.conn.Write(msg.Serialize())
	if err != nil {
		fmt.Println("Error writing request:", err)
	}
}
