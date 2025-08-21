package peers

import (
	"fmt"
	"sync"

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

	var wg sync.WaitGroup

	for i := range pm.peers {
		p := &pm.peers[i]

		if err := p.connect(); err != nil {
			fmt.Println("Error connecting to peer:", err)
			continue
		}

		if err = hs.ExchangeHandshake(p.conn); err != nil {
			p.conn.Close()
			fmt.Println("Handshake failed:", err)
			continue
		}

		if err := pm.AddPeer(p); err != nil {
			fmt.Printf("Error adding peer %s: %v\n", p.IP.String(), err)
			p.conn.Close()
			continue
		}

		wg.Add(1)
		go func(p *Peer) {
			defer wg.Done()
			if err := p.ReadLoop(); err != nil {
				fmt.Printf("Error reading from peer %s: %v\n", p.IP.String(), err)
				p.conn.Close()
			}
		}(p)
	}

	wg.Wait()
}

func (pm *PeerManager) AddPeer(p *Peer) error {
	if p.IP == nil || p.Port <= 0 || p.Port > 65535 || p.conn == nil {
		return fmt.Errorf("invalid peer: %v", p)
	}

	// Check if the peer already exists
	for _, existingPeer := range pm.connectedPeers {
		if existingPeer.IP.Equal(p.IP) {
			return fmt.Errorf("peer already exists: %s", p.IP)
		}
	}

	pm.connectedPeers = append(pm.connectedPeers, p)
	return nil
}

func (pm *PeerManager) RemovePeer(p Peer) error {
	for i, existingPeer := range pm.peers {
		if existingPeer.IP.Equal(p.IP) {
			if existingPeer.conn != nil {
				existingPeer.conn.Close()
			}
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
