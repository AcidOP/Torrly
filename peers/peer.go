package peers

import (
	"net"
	"strconv"
	"time"
)

type Peer struct {
	IP     net.IP
	Port   int
	PeerId string // (Optional) ID of the Peer
	Choked bool
}

type PeerManager struct {
	Peers          []Peer
	ConnectedPeers []Peer
}

func NewPeerManager() *PeerManager {
	peers := []Peer{}
	return &PeerManager{Peers: peers}
}

// COnnect to the associated peer using its IP and Port.
// Returns a net.Conn if successful, or an error if it fails.
// Returned `net.Conn` MUST be closed later by the caller.
func (p Peer) ConnectToPeer() (net.Conn, error) {
	addr := net.JoinHostPort(p.IP.String(), strconv.Itoa(p.Port))
	return net.DialTimeout("tcp", addr, 5*time.Second)
}
