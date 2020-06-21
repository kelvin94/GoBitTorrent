package client

import(
	bf "kelvin94/btClient/biffield"
	"net"
	"fmt"
	peer "kelvin94/btClient/peer"
	msg "kelvin94/btClient/messages"
)


// 		do initTCPConn, handshake, get bitfield message, encapsulate bitField payload, connection, my own peer_id, Choked(boolean)
// A Client is a TCP connection with a peer
type Client struct {
	
	Conn     net.Conn
	Choked   bool
	Bitfield bf.Bitfield
	peer     peer.Peer
	infoHash [20]byte
	peerID   [20]byte
}

func New(peer peer.Peer, peerID [20]byte, infoHash [20]byte) (*Client, error) {
	// finish handshake
	conn, err:=peer.InitTCPConn(infoHash, peerID)
	if err != nil {
		fmt.Println("###client.New()1",err)
		return nil, err
	} 
	// read the stream for bitField message
	bitfield,err := msg.RecvBifFieldsMessage(conn) // RecvMessage returns bitfield message payload only
	if err != nil {
		fmt.Println("###client.New()2",err)
		conn.Close()
		return nil, err
	}
	// 1 byte = 11111111,
	// What is higher bits? example: 0111, "0" is higher bits
	// bitfield message paylod = [255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 240]
	// each byte, such as one of the "255", is equal to binary form "11111111", one bit of the binary form = the picece the the peer has
	// What is the purpose of the bitfield? => The purpose is to tell what pieces the client has, so the other peer know what pieces it can request.


	return &Client{
		Conn : conn,
		Choked : true,
		Bitfield : bitfield,
		peer  : peer,
		infoHash : infoHash,
		peerID : peerID,
	}, nil
}

func (client *Client) SendUnchokeMsg() error {
	msgToBeSent := msg.CreateUnchokeMessage()
	_, err := client.Conn.Write(msgToBeSent.Serialize())
	return err
}

func (client *Client) SendInterestedMsg() error {
	msgToBeSent := msg.CreateInterestedMessage()
	_, err := client.Conn.Write(msgToBeSent.Serialize())
	return err
}

func (client *Client) SendRequestMsg(index int, begin int, length int) error {
	msgToBeSent := msg.CreateRequestMessage(index, begin, length)
	fmt.Println("Sending Request Msg index: ", index, " begin", begin)
	_, err := client.Conn.Write(msgToBeSent.Serialize())
	return err
}

func (client *Client) SendHaveMsg(pieceIndex int) error {
	msgToBeSent := msg.CreateHaveMessage(pieceIndex)
	_, err := client.Conn.Write(msgToBeSent.Serialize())

	return err
}

