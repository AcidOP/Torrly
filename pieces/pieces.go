package pieces

import (
	"bytes"
	"crypto/sha1"
	"fmt"
)

type hash = [20]byte

type Piece struct {
	index      int
	hash       hash // SHA-1 hash of the piece
	length     int  // The length of the piece in bytes
	downloaded int  // Number of bytes already downloaded
	SubPieces  map[int][]byte
	Verified   bool
}

func (p *Piece) AddSubPiece(begin int, subPiece []byte) error {
	if begin < 0 || begin+len(subPiece) > p.length {
		return fmt.Errorf("block out of index")
	}

	if _, existw := p.SubPieces[begin]; existw {
		return fmt.Errorf("subpiece already exists at offset: %d", begin)
	}

	p.SubPieces[begin] = subPiece
	p.downloaded += len(subPiece)

	return nil
}

func (p *Piece) Serialize() []byte {
	buffer := make([]byte, p.length)

	for begin, subPiece := range p.SubPieces {
		copy(buffer[begin:], subPiece)
	}

	return buffer
}

func (p *Piece) Verify() bool {
	if !p.IsComplete() {
		return false
	}

	piece := p.Serialize()
	pieceHash := sha1.Sum(piece)

	p.Verified = bytes.Equal(pieceHash[:], p.hash[:])
	return p.Verified
}

func (p *Piece) IsComplete() bool {
	return p.downloaded == p.length
}
