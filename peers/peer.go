package peers

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/AcidOP/torrly/handshake"
	"github.com/AcidOP/torrly/messages"
)

type Peer struct {
	IP       net.IP
	Port     int
	Choked   bool
	PeerId   string // (Optional)
	conn     net.Conn
	Bitfield []byte
}

type PeerManager struct {
	peers    []Peer
	infoHash []byte
	peerId   []byte
	// connectedPeers []*Peer
}

func NewPeerManager(peers []Peer, infoHash, peerId []byte) *PeerManager {
	return &PeerManager{peers: peers, infoHash: infoHash, peerId: peerId}
}

func (pm *PeerManager) HandlePeers() {
	hs, err := handshake.NewHandshake(pm.infoHash, pm.peerId)
	if err != nil {
		fmt.Println("Error creating handshake:", err)
		return
	}

	for i := range pm.peers {
		p := &pm.peers[i]

		conn, err := p.connect()
		if err != nil || conn == nil {
			continue
		}

		pHandshake, err := hs.ExchangeHandshake(conn)
		if err != nil || len(pHandshake.String()) == 0 {
			fmt.Printf("\nFailed to exchange handshake with peer: %s\n", p.IP.String())
			conn.Close()
			continue
		}

		if err := hs.VerifyHandshake(pHandshake); err != nil {
			continue // Ignore peers that do not match the handshake
		}

		p.conn = conn // Reuse the connection for further communication

		bf, err := p.receiveBitField()
		if err != nil {
			fmt.Println(err)
		} else {
			p.Bitfield = bf
		}

		// pm.connectedPeers = append(pm.connectedPeers, p)
		p.sendUnchoke()
		p.sendInterested()
		p.exchangeMessages()
	}
}

func (p *Peer) Read() (*messages.Message, error) {
	msg, err := messages.Receive(p.conn)
	return msg, err
}

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

	fmt.Printf("\nReceived Bitfield: %x\n\n", msg.Payload)

	return msg.Payload, nil
}

func (p *Peer) sendUnchoke() error {
	msg := messages.Message{ID: messages.MsgUnchoke}
	_, err := p.conn.Write(msg.Serialize())

	fmt.Println("Sent `unchoke` message to peer:", p.IP.String())
	return err
}

func (p *Peer) sendInterested() error {
	msg := messages.Message{ID: messages.MsgInterested}
	_, err := p.conn.Write(msg.Serialize())

	fmt.Println("Sent `ineterested` message to peer:", p.IP.String())
	return err
}

func (p *Peer) exchangeMessages() {
	fmt.Println("\n\nStarting communication with peer: ", p.IP.String())

	for {
		msg, err := messages.Receive(p.conn)
		if err != nil {
			fmt.Println("Error receiving message from peer:", err)
			p.conn.Close()
			return
		}

		if msg.ID == messages.MsgCancel {
			fmt.Println("Received cancel message from peer:", p.IP.String())
			p.conn.Close()
			return
		}
	}
}

// COnnect to the associated peer using its IP and Port.
// Returns a net.Conn if successful, or an error if it fails.
// Returned `net.Conn` MUST be closed later by the caller.
func (p *Peer) connect() (net.Conn, error) {
	addr := net.JoinHostPort(p.IP.String(), strconv.Itoa(p.Port))
	timeout := time.Second * 5
	return net.DialTimeout("tcp", addr, timeout)
}
