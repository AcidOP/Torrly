package main

import (
	"github.com/AcidOP/torrly/torrent"
)

func main() {
	t1, err := torrent.NewTorrentFromFile("./debian.torrent")
	if err != nil {
		panic(err)
	}

	t1.ViewTorrent()
	t1.StartDownload()
}
