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
	connectedPeers []*Peer
}

func NewPeerManager(peers []Peer, infoHash, peerId []byte) *PeerManager {
	return &PeerManager{
		peers:    peers,
		infoHash: infoHash,
		peerId:   peerId,
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

		if err = hs.ExchangeHandshake(conn); err != nil {
			conn.Close()
			continue
		}

		p.conn = conn // Reuse the connection for further communication
		if err := pm.AddPeer(*p); err != nil {
			fmt.Printf("Error adding peer %s: %v\n", p.IP.String(), err)
			conn.Close()
			continue
		}

		go p.ReadLoop() // Start reading messages from the peer
	}
}

func (pm *PeerManager) AddPeer(p Peer) error {
	if p.IP == nil || p.Port <= 0 || p.Port > 65535 || p.conn == nil {
		return fmt.Errorf("invalid peer: %v", p)
	}

	// Check if the peer already exists
	for _, existingPeer := range pm.connectedPeers {
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
		if err := peer.send(msg); err != nil {
			fmt.Printf("Error sending message to peer %s: %v\n", peer.IP.String(), err)
			continue
		}

		fmt.Printf("Broadcasted message (%s) to peer: %s\n", msg.String(), peer.IP.String())
	}
}
