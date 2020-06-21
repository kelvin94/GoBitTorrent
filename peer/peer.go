package peer

import(
	"net"
	"fmt"
	"strconv"
	"encoding/binary"
	// "net"
	"time"
	// "bufio"
	"bytes"
	// "errors"
	"io"
	"log"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

type HandShake struct {
	Pstr     string
	Info_hash [20]byte
	Peer_id [20]byte
}

func UnmarshalPeers(data string) []Peer{

	// each peer is made of 6 bytes of the data
	//		4 bytes = IP
	//		2 bytes = Port, uint16
	numOfBytesPerPeer := 6
	numOfPeers := len(data) / numOfBytesPerPeer
	var peers []Peer
	for i := 0; i < numOfPeers; i++ {
		// temp := Peer{}
		start := i*6 // chop the line of raw bytes into slices
		end := (i+1)*6
		slice := data[start:end]
		var ip string = slice[0:4]
		var port string = slice[4:]
		
		// LEARNED: binary.BigEndian.Uint16() - convert an integer which is represented by 16-bits
		peers = append(peers,Peer{net.IP([]byte(ip)), binary.BigEndian.Uint16([]byte(port))})

	}
	
	return peers
}

func (p *Peer) InitTCPConn(Info_hash [20]byte, peer_id [20]byte) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", p.ToString(), 3 * time.Second)
	// fmt.Println("infohash", *infohash)
	if err != nil {
		log.Println("InitTCPConn connection err", err)
		return nil,err
	}
	h := NewHandShake(Info_hash, peer_id)
	receivedHandShake, err := h.doHandShake(conn,Info_hash, peer_id)
	if err == nil {
		if !bytes.Equal((*receivedHandShake).Info_hash[:], Info_hash[:]) {
			return nil,fmt.Errorf("Expected infohash %x but got %x", (*receivedHandShake).Info_hash, Info_hash)
		}

	}
	
	return conn,nil
}

func (h *HandShake) doHandShake(conn net.Conn, infohash [20]byte, peerId [20]byte) (*HandShake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	// defer conn.Close()
	// Peer wire protocol implementation
	var serializedHandShake []byte = h.Serialize()
	
	
	_, err := conn.Write(serializedHandShake)
	if err != nil {
		log.Println("doHandShake err", err)
		return nil,err
	}
	receivedHandShake,err := ParseTCPMessage(conn)
	if err != nil {
		return nil, err
	}
	return receivedHandShake,nil
}

func NewHandShake( infohash [20]byte, peerID [20]byte) *HandShake {
	return &HandShake{
		Pstr:     "BitTorrent protocol",
		Info_hash: infohash,
		Peer_id:   peerID,
	}
}

// Serialize serializes a message into a buffer of the form
// <length prefix><message ID><payload>
// Interprets `nil` as a keep-alive message
// Refernece: Jesse Li blog post
func (h *HandShake) Serialize() []byte {
	// NOTE: Whenever we are sending a msg, it has to be in []byte type 
	buf := make([]byte, len(h.Pstr)+49)
	buf[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(buf[curr:], h.Pstr)
	curr += copy(buf[curr:], make([]byte, 8)) // 8 reserved bytes
	curr += copy(buf[curr:], h.Info_hash[:])
	curr += copy(buf[curr:], h.Peer_id[:])
	return buf
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func ParseTCPMessage(r io.Reader) (*HandShake, error) {
	pstrLengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, pstrLengthBuf)
	checkError(err)
	pstrlen := int(pstrLengthBuf[0]) // cast byte to int
	if pstrlen <= 0 {
		err := fmt.Errorf("pstrlen cannot be <= 0")
		return nil,err
	}

	remainingMsgBuf := make([]byte, 48+pstrlen)
	
	_, err = io.ReadFull(r, remainingMsgBuf)
	if err != nil {
		return nil, err
	}
	
	var infoHash, peerId [20]byte
	copy(infoHash[:], remainingMsgBuf[8+pstrlen:8+20+pstrlen])
	copy(peerId[:],remainingMsgBuf[8+pstrlen+20:])
	// conn.Read(buf)
	h := &HandShake{
		Pstr:     string(remainingMsgBuf[0:pstrlen]),
		Info_hash: infoHash,
		Peer_id:   peerId,
	}
	return h,nil
}


func (p *Peer) ToString() (string) {
	var ip *net.IP = &p.IP
	var port *uint16 = &p.Port
	return net.JoinHostPort((*ip).String(),strconv.FormatInt(int64(*port), 10))
}

