package messages
import(
	"encoding/binary"
	"time"
	"net"
	"io"
	// "bytes"
	"fmt"
	bf "kelvin94/btClient/biffield"
)

const(
	// MsgChoke chokes the receiver
	MsgIDChoke messageID = 0
	// MsgUnchoke unchokes the receiver
	MsgIDUnchoke messageID = 1
	// MsgInterested expresses interest in receiving data
	MsgIDInterested messageID = 2
	// MsgNotInterested expresses disinterest in receiving data
	MsgIDNotInterested messageID = 3
	// MsgHave alerts the receiver that the sender has downloaded a piece
	MsgIDHave messageID = 4
	// MsgBitfield encodes which pieces that the sender has downloaded
	MsgIDBitField messageID = 5
	// MsgRequest requests a block of data from the receiver
	MsgIDRequest messageID = 6
	// MsgPiece delivers a block of data to fulfill a request
	MsgIDPiece messageID = 7
	// MsgCancel cancels a request
	MsgIDCancel messageID = 8
)

type messageID uint8

/**
The length-prefix is a four-byte big-endian value. 
The messageID is a single decimal byte. 
The payload is message dependent.

Reference: https://www.cs.rochester.edu/courses/257/fall2017/projects/gp.html
***/
type Message struct {
	MessageID messageID
	Payload []byte // The payload is a bitfield representing the pieces that have been successfully downloaded. 
}




func ReadChokeUnchokeMessage(r io.Reader) (*Message, error) {
	var buf []byte
	io.ReadFull( r, buf)
	/**
	The length-prefix is a four-byte big-endian value. Length-prefix = length of the whole message
	The messageID is a single decimal byte. 
	The payload is message dependent.

	Reference: https://www.cs.rochester.edu/courses/257/fall2017/projects/gp.html
	***/
	lengthPrefixBuf := make([]byte, 4) // The length prefix is a four byte big-endian value.
	_, err := io.ReadFull(r, lengthPrefixBuf)
	if err != nil {
		// fmt.Println("###message.readChokeUnchokeMessage() io.ReadFull1 error", err)
		return nil,err
	}
	length := binary.BigEndian.Uint32(lengthPrefixBuf)

	// keep-alive message
	if length == 0 {
		return  nil,nil
	}
	messageBuf := make([]byte, length)
	
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		// fmt.Println("###message.readBitFields() io.ReadFull2 error", err)
		return nil,err
	}

	m := &Message{
		MessageID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}
	if m.MessageID == MsgIDUnchoke {
		fmt.Println("**********Yes!! Unchoked from the peer")
	}
	return m,nil
}

func readBitFields(r io.Reader) (*Message, error) {
	var buf []byte
    io.ReadFull( r, buf)
	lengthPrefixBuf := make([]byte, 4) // The length prefix is a four byte big-endian value.
	_, err := io.ReadFull(r, lengthPrefixBuf)
	if err != nil {
		fmt.Println("###message.readBitFields() io.ReadFull1 error", err)
		return nil,err
	}
	length := binary.BigEndian.Uint32(lengthPrefixBuf)
	// keep-alive message
	if length == 0 {
		return  nil,nil
	}

	messageBuf := make([]byte, length)
	
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		fmt.Println("###message.readBitFields() io.ReadFull2 error", err)
		return nil,err
	}

	m := &Message{
		MessageID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}
	return m,nil
}


// msg format: <length prefix> <message ID> <payload>
// If the message length isn’t greater than 4, then we know it is the keep-ahead message which has no id. 
//If the length isn’t greater than 5 we know that it has no payload. 
//If the id is 6, 7, 8, those messages split the payload into index, begin, and block/length.
// reference: https://allenkim67.github.io/programming/2016/05/04/how-to-make-your-own-bittorrent-client.html#implementing-the-message-handlers
func (msg *Message) Serialize() []byte {
	// NOTE: Whenever we are sending a msg, it has to be in []byte type 
	var wholeMsgLength uint32 = uint32(4+1+len(msg.Payload))
	buf := make([]byte, wholeMsgLength) // "1+" is for the message ID, "4" is for the length-prefix
	
	/**
	The length-prefix is a four-byte big-endian value. Length-prefix = length of the whole message
	The messageID is a single decimal byte. 
	The payload is message dependent.

	Reference: https://www.cs.rochester.edu/courses/257/fall2017/projects/gp.html
	***/
	
	// we need to insert the byte of an 32bits integer into the first 4 bytes of the buffer
	binary.BigEndian.PutUint32(buf, wholeMsgLength) 
	buf[4] = byte(msg.MessageID)
	copy(buf[5:], msg.Payload)
	return buf
}

func CreateUnchokeMessage() (Message) {
	// Unchoke Message has no payload
	// reference: https://www.cs.rochester.edu/courses/257/fall2017/projects/gp.html
	return Message{
		MessageID: MsgIDUnchoke,
	}
}

func CreateInterestedMessage() (Message) {
	// Unchoke Message has no payload
	// reference: https://www.cs.rochester.edu/courses/257/fall2017/projects/gp.html
	return Message{
		MessageID: MsgIDInterested,
	}
}


func CreateHaveMessage(pieceIndex int) (Message) {
	// Unchoke Message has no payload
	// reference: https://www.cs.rochester.edu/courses/257/fall2017/projects/gp.html
	// Have message payload: 4-byte integer -> 32 bit integer
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(pieceIndex))
	return Message{
		MessageID: MsgIDHave,
		Payload: payload,
	}
}
/**
The request message is fixed length, and is used to request a block. 
The payload concatenates index, begin, and length. 
The length of the payload is 12 byte (3 integers)
*/
func CreateRequestMessage(index int, begin int, length int ) (Message) {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint32(buf[0:4], uint32(index))
	binary.BigEndian.PutUint32(buf[4:8], uint32(begin))
	binary.BigEndian.PutUint32(buf[8:12], uint32(length))
	
	return Message{
		MessageID: MsgIDRequest,
		Payload: buf,
	}
}

func RecvBifFieldsMessage(conn net.Conn) (bf.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline
	
	msg, err := readBitFields(conn)
	if err != nil {
		conn.Close()
		fmt.Println("###RecvMessage error", err)
		return nil,err
	}
	if (*msg).MessageID != MsgIDBitField {
		err := fmt.Errorf("Expected bitfield but got ID %d", msg.MessageID) // LEARNED: The bitfield message may only be sent immediately after the handshaking sequence is completed, and before any other messages are sent. The bitfield needs to be sent if a client has pieces to share.
		return nil,err
	}

	return msg.Payload,nil
}

/**
	Process the "Piece" message, which is sent by peers, containing data that we have  requested
***/
func ProcessPiece(expectedPieceIndex int, buf []byte, msg *Message) (int, error) {

	payload := msg.Payload
	if msg.MessageID != MsgIDPiece {
		return 0, fmt.Errorf("Expected PIECE (ID %d), got ID %d", MsgIDPiece, msg.MessageID)
	}
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("Payload too short. %d < 8", len(msg.Payload))
	}
	
	index := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if expectedPieceIndex != index {
		return 0, fmt.Errorf("Expected index %d, got %d", expectedPieceIndex, index)
	}

	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("Begin offset too high. %d >= %d", begin, len(buf))
	}

	blockOfData := payload[8:]
	if begin+len(blockOfData) > len(buf) {
		return 0, fmt.Errorf("Data too long [%d] for offset %d with length %d", len(blockOfData), begin, len(buf))
	}

	/** 
		Lesson learned
		fmt.Println("len(buf) ", len(buf))
		fmt.Println("len(blockOfData) ", len(blockOfData))
		copy(buf[begin:len(blockOfData)], blockOfData) // FIXME: This is not right. len(blockOfData) is always gonna be 16384(16kb) which we ask the peer to send to us, so begin will be larger than 16384 after we receive first block of data from peer
	*/
	copy(buf[begin:], blockOfData)


	
	return len(blockOfData), nil

}