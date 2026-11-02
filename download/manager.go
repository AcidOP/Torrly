package download

import (
	"github.com/AcidOP/torrly/peers"
	"github.com/AcidOP/torrly/pieces"
	"github.com/AcidOP/torrly/torrent"
)

type DownloadManager struct {
	torrent      *torrent.Torrent
	peerManager  *peers.PeerManager
	pieceManager *pieces.PieceManager
	// storageManager  *storage.StorageManager
}

func New(t *torrent.Torrent, peerM *peers.PeerManager, pieceM *pieces.PieceManager) *DownloadManager {
	return &DownloadManager{
		torrent:      t,
		peerManager:  peerM,
		pieceManager: pieceM,
	}
}
