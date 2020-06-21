package torrentfile

import(
	// "crypto/sha1"
	bencode "github.com/jackpal/bencode-go"
	peer "kelvin94/btClient/peer"

	"bytes"
	"crypto/rand"
	"net/http"
	"time"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"errors"
)

type TrackerRequiredInfo struct {
	info_hash [20]byte
	peer_id [20]byte // length 20
	port int
}

type TrackerResponse struct {
	Interval int `bencode:"interval"`
	// Complete int `bencode:"complete"`
	// Incomplete int `bencode:"incomplete"`
	Peers    string `bencode:"peers"`
}

type announceResponse struct {
	// FailureReason  string             `bencode:"failure reason"`
	// RetryIn        string             `bencode:"retry in"`
	// WarningMessage string             `bencode:"warning message"`
	Interval       int32              `bencode:"interval"`
	// MinInterval    int32              `bencode:"min interval"`
	// TrackerID      string             `bencode:"tracker id"`
	// Complete       int32              `bencode:"complete"`
	// Incomplete     int32              `bencode:"incomplete"`
	Peers          string `bencode:"peers"`
	// ExternalIP     []byte             `bencode:"external ip"`
}

// Expect: a string url to make GET request via curl
func (t *TorrentFile) BuildTrackerGETRequest() (string, [20]byte, error) {
	trackerRequiredInfo := TrackerRequiredInfo{}
	
	trackerRequiredInfo.info_hash = t.InfoHash
	
	peerid, err := generateClientPeerId()
	if err != nil {
		fmt.Println("###BuildTrackerGETRequest generateClientPeerId() eeror,",err)
		return "", [20]byte{}, errors.New("fail to BuildTrackerGETRequest, message")
	}
	trackerRequiredInfo.peer_id = peerid
	trackerRequiredInfo.port =  6881 


	req, err := url.Parse(t.Announce)
	if err != nil {
		fmt.Println(err)
	}
	

	q := url.Values{
		"info_hash":  []string{string(trackerRequiredInfo.info_hash[:])},
		"peer_id":    []string{string(trackerRequiredInfo.peer_id[:])},
		"port":       []string{strconv.Itoa(trackerRequiredInfo.port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Info.Length)},
	}
		
	req.RawQuery = q.Encode() // encode() to do urlencoding	
	return req.String(), trackerRequiredInfo.peer_id, nil
}




/**
	generateClientPeerId(): generate the peer id for myself
***/
func generateClientPeerId() ([20]byte, error) {
	var peerID [20]byte
	// Read() generates len(p) random bytes and writes them into p. It always returns len(p) and a nil error.
	_, err := rand.Read(peerID[:])
	return peerID, err	
}

func (t *TorrentFile) CallTracker(url string) ([]peer.Peer, error){
	c := &http.Client{Timeout: 30 * time.Second}
	res, err := c.Get(url)
	if err != nil {
		fmt.Println("###CallTracker() error", err)
		return nil, errors.New("CallTracker() error")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		fmt.Println("tracker get request problem")
		return nil, errors.New("tracker get request problem")
	}
	// var responseBody []byte // LEARNED: in golang, uint8 is "ALIAS OF byte"
	// responseBody, err = ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	var buf bytes.Buffer
	io.Copy(&buf, res.Body)
	tr := announceResponse{}
	eors := bencode.Unmarshal(bytes.NewReader(buf.Bytes()), &tr)
	if eors != nil {
		fmt.Println("###CallTracker() bencode.Unmarshal", eors)
		return nil, errors.New("CallTracker() bencode.Unmarshal")

	}
	return peer.UnmarshalPeers(tr.Peers), nil
}


func typeof(v interface{}) string {
    return fmt.Sprintf("%T", v)
}