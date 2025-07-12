package peers

import (
	"net"
	"strconv"
	"time"
)

type Peer struct {
	IP     net.IP // IP address of the peer in binary format.
	Port   int    // Port number of the peer to connect to.
	PeerId string // (Optional) ID of the Peer
}

// COnnect to the associated peer using its IP and Port.
// Returns a net.Conn if successful, or an error if it fails.
// Returned `net.Conn` MUST be closed later by the caller.
func (p Peer) ConnectToPeer() (net.Conn, error) {
	addr := net.JoinHostPort(p.IP.String(), strconv.Itoa(p.Port))
	return net.DialTimeout("tcp", addr, 5*time.Second)
}
