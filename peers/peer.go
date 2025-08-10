package peers

import (
	"fmt"
	"net"
	"strconv"
	"time"

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
				fmt.Printf("\n[%s] Expected piece message, but got ID %d\n", p.IP.String(), msg.ID)
				continue
			}

			fmt.Printf("\n[%s] Received %d bytes for piece %d\n", p.IP.String(), len(msg.Payload), i)
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

func (p *Peer) hasBitfield(index int) bool {
	if index < 0 || index >= len(p.Bitfield)*8 {
		return false
	}

	byteIndex := index / 8
	bitIndex := index % 8
	return (p.Bitfield[byteIndex] & (1 << (7 - bitIndex))) != 0
}
