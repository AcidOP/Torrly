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
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{t.PeerId},
		"port":       []string{strconv.Itoa(t.Port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{strconv.Itoa(t.Length)},
	}

	base.RawQuery = params.Encode()
	return base.String(), nil
}

// Announce to the tracker to get a list of peers
// Returns a map of peers with their IP addresses and ports
func getTrackerResponse(t Torrent) ([]byte, error) {
	trackerURL, err := t.buildTrackerURL()
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(trackerURL)
	if err != nil {
		return nil, errors.New("failed to connect to tracker")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("tracker returned non-200 status code: " + resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed to read response: " + err.Error())
	}
	return data, nil
}

// Returns a list of peers from the tracker
func (t Torrent) FetchPeers() ([]peers.Peer, error) {
	res, err := getTrackerResponse(t)
	if err != nil {
		return nil, err
	}

	tr := TrackerResponse{}
	if err = bencode.Unmarshal(bytes.NewReader(res), &tr); err != nil {
		return nil, err
	}

	pArr := []peers.Peer{}
	for _, p := range tr.Peers {
		ip := net.ParseIP(p.IP)

		// Filter out malformed Peers
		if ip.String() == "<nil>" || p.Port == 0 {
			continue
		}

		pArr = append(pArr, peers.Peer{
			IP:     ip,
			Port:   p.Port,
			PeerId: p.PeerId,
		})
	}
	return pArr, nil
}
