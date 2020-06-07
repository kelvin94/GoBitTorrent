package torrentfile

import(
	"crypto/sha1"
	bencode "github.com/jackpal/bencode-go"
	"bytes"
	"fmt"
	"crypto/rand"
	"net/http"
	// "net/url"
	// "strconv"
)

type TrackerInfo struct {

}

type TrackerRequiredInfo struct {
	info_hash [20]byte
	peer_id [20]byte // length 20
	port int
}

type TrackerResponse struct {

}

// Expect: a string url to make GET request via curl
func (t *TorrentFile) BuildTrackerGETRequest() (string, error) {
	trackerRequiredInfo := TrackerRequiredInfo{}
	trackerRequiredInfo.info_hash = generateInfoHash(t.Info)
	peerid, err := generateClientPeerId()
	if err != nil {
		fmt.Println(err)
	}
	trackerRequiredInfo.peer_id = peerid
	trackerRequiredInfo.port =  6881

	req, err := http.NewRequest("get", t.Announce, nil)
	if err != nil {
		fmt.Println(err)
	}
	
	q := req.URL.Query()
	q.Add("info_hash", string(trackerRequiredInfo.info_hash[:])) // QueryEscape to do urlencoding
	q.Add("peer_id", string(trackerRequiredInfo.peer_id[:]))
	q.Add("port", string(trackerRequiredInfo.port))
	q.Add("uploaded", "0")
	q.Add("downloaded", "0")
	q.Add("compact", "1")
	// req.URL.RawQuery = q.Encode()
	req.URL.RawQuery = q.Encode()
	fmt.Println("url encoded url: ", req.URL.String())
	return req.URL.String(),nil
}


func generateInfoHash(tmi TorrentMetaInfo) [20]byte {
	var b bytes.Buffer // bytes.Buffer struct implements io.Writer interface as well!!!
	fmt.Println("before encoded to bencode, ",tmi)
	bencode.Marshal(&b,tmi)

	h := sha1.Sum(b.Bytes())	
	return h
}

func generateClientPeerId() ([20]byte, error) {
	// the peer id for myself
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	return peerID, err
	
}

