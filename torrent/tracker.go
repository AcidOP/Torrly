package torrent

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/AcidOP/torrly/peers"
	"github.com/jackpal/bencode-go"
)

type peer struct {
	IP     string `bencode:"ip"`
	Port   int    `bencode:"port"`
	PeerId string `bencode:"peer id"`
}

type TrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    []peer `bencode:"peers"`
}

// Create a URL to request to the tracker for peer information
// Must be a GET request with the following:
// https://wiki.theory.org/BitTorrentSpecification#Tracker_Request_Parameters
func (t Torrent) buildTrackerURL() (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  {string(t.InfoHash[:])},
		"peer_id":    {t.PeerId},
		"port":       {strconv.Itoa(t.Port)},
		"uploaded":   {"0"},
		"downloaded": {"0"},
		"left":       {strconv.Itoa(t.Length)},
	}

	base.RawQuery = params.Encode()
	return base.String(), nil
}

// Announce to the tracker to get a list of peers
// Returns a map of peers with their IP addresses and ports
func (t *Torrent) getTrackerResponse() ([]byte, error) {
	trackerURL, err := t.buildTrackerURL()
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(trackerURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed to read tracker response: " + err.Error())
	}
	return data, nil
}

// Returns a list of peers from the tracker
func (t Torrent) GetAvailablePeers() ([]peers.Peer, error) {
	res, err := t.getTrackerResponse()
	if err != nil {
		return nil, err
	}

	tr := TrackerResponse{}
	if err = bencode.Unmarshal(bytes.NewReader(res), &tr); err != nil {
		return nil, err
	}

	pArr := []peers.Peer{}
	for _, p := range tr.Peers {
		pArr = append(pArr, peers.Peer{
			IP:     net.ParseIP(p.IP),
			Port:   p.Port,
			Choked: true,
		})
	}
	return pArr, nil
}
