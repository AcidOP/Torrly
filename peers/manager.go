package peers

import (
	"fmt"

	"github.com/AcidOP/torrly/handshake"
	"github.com/AcidOP/torrly/messages"
)

type PeerManager struct {
	peers          []Peer
	infoHash       []byte
	peerId         []byte
	pieceHashes    []hash
	pieceLength    int
	totalLength    int
	connectedPeers []*Peer
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

		// If the handshake is invalid, we ignore this peer.
		if err := hs.VerifyHandshake(pHandshake); err != nil {
			continue
		}

		p.conn = conn // Reuse the connection for further communication

		msg, err := p.receiveBitField()
		if err != nil {
			fmt.Println(err)
		}

		if msg.ID == messages.MsgBitfield {
			p.Bitfield = msg.Payload
		}

		// First message might not be Bitfield, so we check if it is Unchoke
		if msg.ID == messages.MsgUnchoke {
			p.Choked = false
		} else {
			p.Choked = true
		}

		// pm.connectedPeers = append(pm.connectedPeers, p)
		p.startDownloader(pm.pieceHashes, pm.pieceLength)
	}
}

func (pm *PeerManager) AddPeer(p Peer) error {
	if p.IP == nil || p.Port <= 0 || p.Port > 65535 || p.conn == nil {
		return fmt.Errorf("invalid peer: %v", p)
	}

	// Check if the peer already exists
	for _, existingPeer := range pm.peers {
		if existingPeer.IP.Equal(p.IP) {
			return fmt.Errorf("peer already exists: %s", p.IP)
		}
	}

	pm.peers = append(pm.peers, p)
	return nil
}

func (pm *PeerManager) RemovePeer(p Peer) error {
	for i, existingPeer := range pm.peers {
		if existingPeer.IP.Equal(p.IP) {
			pm.peers = append(pm.peers[:i], pm.peers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("peer not found: %s", p.IP)
}

func (pm *PeerManager) BroadcastMessage(msg *messages.Message) {
	for _, peer := range pm.connectedPeers {
		if peer.Choked {
			continue
		}

		if err := peer.Send(msg); err != nil {
			fmt.Printf("Error sending message to peer %s: %v\n", peer.IP.String(), err)
			continue
		}

		fmt.Printf("Broadcasted message (%s) to peer: %s\n", msg.String(), peer.IP.String())
	}
}
