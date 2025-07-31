package peers

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/AcidOP/torrly/handshake"
	"github.com/AcidOP/torrly/messages"
)

type hash = [20]byte

type Peer struct {
	IP       net.IP
	Port     int
	Choked   bool
	conn     net.Conn
	Bitfield []byte
}

type PeerManager struct {
	peers       []Peer
	infoHash    []byte
	peerId      []byte
	pieceHashes []hash
	pieceLength int
	totalLength int
	// connectedPeers []*Peer
}

func NewPeerManager(
	peers []Peer, infoHash, peerId []byte, pieces []hash, pieceLength, totalLength int,
) *PeerManager {
	return &PeerManager{
		peers:       peers,
		infoHash:    infoHash,
		peerId:      peerId,
		pieceHashes: pieces,
		pieceLength: pieceLength,
		totalLength: totalLength,
	}
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

		p.conn = conn   // Reuse the connection for further communication
		p.Choked = true // Initially choked

		bf, err := p.receiveBitField()
		if err != nil {
			fmt.Println(err)
		} else {
			p.Bitfield = bf
		}

		// pm.connectedPeers = append(pm.connectedPeers, p)
		p.startDownloader(pm.pieceHashes, pm.pieceLength)
	}
}

// Read function reads a `messages.Message` from the peer's connection.
// (Optionally) accepts a timeout duration to set a read deadline.
// If no timeout is provided, it defaults to 5 seconds.
func (p *Peer) Read(timeout ...time.Duration) (*messages.Message, error) {
	duration := 5 * time.Second // Default timeout
	if len(timeout) > 0 {
		duration = timeout[0]
	}

	p.conn.SetReadDeadline(time.Now().Add(duration))
	msg, err := messages.Receive(p.conn)
	return msg, err
}

func (p *Peer) startDownloader(pieces []hash, pieceLength int) {
	fmt.Println("\n\nRequesting pieces from: ", p.IP.String())

	p.sendInterested()

	// Check if the peer choked us
	msg, err := p.Read()
	if err != nil {
		fmt.Println("Error reading from peer:", err)
		return
	}

	if msg.ID == messages.MsgUnchoke {
		p.Choked = false
	}

	for i := range pieces {
		if p.Choked {
			fmt.Println("Peer has choked us, cannot download pieces.")
			return
		}

		fmt.Printf("Requesting piece %d from peer %s\n", i, p.IP.String())

		blockSize := 16 * 1024 // 16 KB
		for begin := 0; begin < pieceLength; begin += blockSize {
			length := blockSize
			if begin+length > pieceLength {
				length = pieceLength - begin
			}

			p.sendRequest(i, length, begin)

			msg, err = p.Read()
			if err != nil {
				fmt.Println("Error reading piece from peer:", err)
				return
			}

			if msg.ID != messages.MsgPiece {
				fmt.Printf("Expected piece message, but got ID %d from peer %s\n", msg.ID, p.IP.String())
				continue
			}

			fmt.Printf("Received %d bytes for piece %d from peer %s\n", len(msg.Payload), i, p.IP.String())
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
