package torrent

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jackpal/bencode-go"
)

type Torrent struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
	PeerId      string // Our own Peer ID, used for handshakes.
	Port        int    // Port we listen on for incoming connections
}

type bInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bTorrent struct {
	Announce string `bencode:"announce"`
	Info     bInfo  `bencode:"info"`
}

const (
	PeerID = "-TRLY01-9a8b7c6d5e4f"
	Port   = 6881
)

func NewTorrentFromFile(path string) (*Torrent, error) {
	f, err := parseTorrentFromPath(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tf, err := metaFromFile(f)
	if err != nil {
		return nil, err
	}
	return &tf, nil
}

// Calculate the SHA1 hash of the bencoded info dictionary.
func (i *bInfo) hash() [20]byte {
	infoBytes := bytes.Buffer{}
	if err := bencode.Marshal(&infoBytes, *i); err != nil {
		panic("failed to marshal info: " + err.Error())
	}

	return sha1.Sum(infoBytes.Bytes())
}

// Visualize information about the torrent.
// (e.g. announce URL, file name, size, piece length, number of pieces, info hash)
func (t *Torrent) ViewTorrent() {
	var displaySize string
	s := t.Length / (1024 * 1024)

	// Format the size in GB, MB, or KB
	if s >= 1024 {
		displaySize = fmt.Sprintf("%.2f GB", float64(s)/1024)
	} else if s >= 1 {
		displaySize = fmt.Sprintf("%.2f MB", float64(s))
	} else {
		displaySize = fmt.Sprintf("%d KB", t.Length/1024)
	}

	fmt.Printf("\n\nAnnounce: %s\n", t.Announce)
	fmt.Printf("File name: %s\n", t.Name)
	fmt.Printf("File size: %s\n", displaySize)
	fmt.Printf("Piece length: %d KB\n", t.PieceLength/1024)
	fmt.Printf("Num pieces: %d\n", t.PieceLength/20)
	fmt.Printf("Info Hash: %x\n\n", t.InfoHash)
}

func (t *Torrent) StartDownload() {
	pArr, err := t.FetchPeers()
	if err != nil {
		panic(err)
	}

	for _, p := range pArr {
		// err := peers.HandshakePeer(p, string(t.InfoHash[:]), t.PeerId)
		// if err != nil {
		// 	fmt.Println("Failed to handshake with peer:", p.IP, "on port", p.Port, "-", err)
		// 	continue
		// }
		p.HandshakePeer(string(t.InfoHash[:]), t.PeerId)
	}
}

// Takes a path as an argument and checks if the file is a .torrent file.
// Then reads the file and a pointer to the file.
func parseTorrentFromPath(path string) (*os.File, error) {
	f, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}

	// Check the extension
	if strings.Split(f.Name(), ".")[1] != "torrent" {
		return nil, errors.New("file " + f.Name() + " is not a .torrent file")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Takes a file as argument and reads the torrent metadata from it.
// Returns a Torrent struct with the metadata.
func metaFromFile(f *os.File) (Torrent, error) {
	if f == nil {
		return Torrent{}, errors.New("file pointer is nil")
	}

	bt := bTorrent{}
	if err := bencode.Unmarshal(f, &bt); err != nil {
		return Torrent{}, errors.New("failed to parse torrent file: " + err.Error())
	}

	// SHA1 hash of `info` dictionary
	iHash := bt.Info.hash()

	// Split the pieces into an array of  hashes
	pHashes, err := splitPieceHashes(bt.Info)
	if err != nil {
		return Torrent{}, err
	}

	t := Torrent{
		Announce:    bt.Announce,
		InfoHash:    iHash,
		PieceHashes: pHashes,
		PieceLength: bt.Info.PieceLength,
		Length:      bt.Info.Length,
		Name:        bt.Info.Name,
		PeerId:      PeerID,
		Port:        Port,
	}
	return t, nil
}

// Take the `info` key from meta and split the pieces into an array of hashes.
// Returns an array of 20-byte hashes.
func splitPieceHashes(i bInfo) ([][20]byte, error) {
	hashLen := 20 // SHA1 is 20 bytes long

	buf := make([]byte, len(i.Pieces))
	if len(buf)%hashLen != 0 {
		return nil, errors.New("malformed pieces: " + fmt.Sprint(len(buf)))
	}

	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := range numHashes {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}
