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
	IP     net.IP
	Port   int
	Choked bool
	PeerId string // (Optional)
	conn   net.Conn
}

type PeerManager struct {
	peers          []Peer
	infoHash       string
	peerId         string
	connectedPeers []*Peer
}

func NewPeerManager(peers []Peer, infoHash, peerId string) *PeerManager {
	return &PeerManager{peers: peers, infoHash: infoHash, peerId: peerId}
}

func (pm *PeerManager) HandlePeers() {
	hs := handshake.NewHandshake(pm.infoHash, pm.peerId)

	for i := range pm.peers {
		p := &pm.peers[i]

		conn, err := p.ConnectToPeer()
		if err != nil || conn == nil {
			continue
		}

		pHandshake, err := hs.ExchangeHandshake(conn)
		if err != nil || pHandshake == nil {
			fmt.Printf("\nFailed to exchange handshake with peer: %s", p.IP.String())
			conn.Close()
			continue
		}

		matched := hs.VerifyHandshake(pHandshake)

		if matched {
			pm.connectedPeers = append(pm.connectedPeers, p)
			p.conn = conn
			p.exchangeMessages()
		}
	}
}

func (p Peer) exchangeMessages() {
	fmt.Println("\n\nStarting communication with peer: ", p.IP.String())

	messages.ReceivePeerMessage(p.conn)
	p.conn.Close()
}

// COnnect to the associated peer using its IP and Port.
// Returns a net.Conn if successful, or an error if it fails.
// Returned `net.Conn` MUST be closed later by the caller.
func (p Peer) ConnectToPeer() (net.Conn, error) {
	addr := net.JoinHostPort(p.IP.String(), strconv.Itoa(p.Port))
	return net.DialTimeout("tcp", addr, 5*time.Second)
}
