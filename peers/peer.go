package peers

import (
	"errors"
	"net"
	"strconv"
	"time"
)

type Peer struct {
	IP     net.IP // IP address of the peer in binary format.
	Port   int    // Port number of the peer to connect to.
	PeerId string // (Optional) ID of the Peer
}

// Try to establish a connection to the peer.
// Returns a net.Conn if successful, or an error if it fails.
// Returned `net.Conn` MUST be closed later by the caller.
func (p Peer) ConnPeer() (net.Conn, error) {
	addr := net.JoinHostPort(p.IP.String(), strconv.Itoa(p.Port))

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, err
	}

	if conn == nil {
		return nil, errors.New("failed to connect to peer: " + addr)
	}
	return conn, nil
}
