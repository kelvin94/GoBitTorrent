package downloader

import(
	peer "kelvin94/btClient/peer"
	client "kelvin94/btClient/client"
	"fmt"
	// "sync"
	"time"
	// "errors"
	msg "kelvin94/btClient/messages"
	// "net"
	"log"
	"runtime"
	"bytes"
	"crypto/sha1"

)
type Torrent struct {
	Peers []peer.Peer
	InfoHash [20]byte
	PeerId [20]byte
	PieceLength int
	Pieces [][20]byte
	Length int
}

type workmsg struct {
	pieceIndex int
	lengthOfPiece int
	piece [20]byte
}

type datamsg struct {
	pieceIndex int
	data []byte
}


type stateProgress struct {
	index int
	requestedByte int // primititive types in golang will be initilized to 0
	requestedBlock int // primititive types in golang will be initilized to 0
	numOfPipelinedRequests int
	downloaded int // primititive types in golang will be initilized to 0
	c *client.Client
	buf []byte // buffer that stores the bytes that have been downloaded
}

/***
	Find out the start index and end index for a piece of data
***/
func calculateStartAndEndOfAPiece(index int, pieceLength int, fileTotalLength int) (begin int, end int) {
	// Why we need to calculate
	fmt.Println("calculateStartAndEndOfAPiece index ", index, " pieceLength", pieceLength)
	begin = index * pieceLength
	end = begin + pieceLength
	if end > fileTotalLength {
		end = fileTotalLength
	}
	return begin, end
}

func calculatePieceLength(index int, pieceLength int, fileTotalLength int) int {
	begin, end := calculateStartAndEndOfAPiece(index, pieceLength, fileTotalLength)
	return end - begin
}

func (t *Torrent) Download() ([]byte, error) {
	workqueue := make(chan *workmsg, len(t.Pieces))
	dataqueue := make(chan *datamsg)
	// for each piece, create a workmsg and dump it to the workqueue.
	// pre-loaded the queue of jobs
	for i, hash := range(t.Pieces) {
		workmsg := &workmsg{
			pieceIndex: i,
			lengthOfPiece: calculatePieceLength(i, t.PieceLength, t.Length),
			piece: hash,
		}
		fmt.Println("workmsg going into the queue", workmsg)
		workqueue <- workmsg
	}
	for _, peeer := range t.Peers {

			go t.spawnWorker(peeer, workqueue, dataqueue)
		
		
	}
	buf := make([]byte, t.Length)
	count := 0

	// ##NOTE-1
	for count < len(t.Pieces) {
		data := <- dataqueue
		start, end := calculateStartAndEndOfAPiece(data.pieceIndex, t.PieceLength, t.Length)
		fmt.Println("###GOt the piece piece starting at the index", start)
		copy(buf[start:end], data.data)
		count++
		fmt.Println("###Checking pieces #%d integrity...", data.pieceIndex)

		fmt.Println("###pieces #%d integrity(True/False): %s", checkIntegrity(data.data, t.Pieces[data.pieceIndex]))

		/**
			Reference: https://blog.jse.li/posts/torrent/
		***/
		percent := float64(count) / float64(len(t.Pieces)) * 100
		numWorkers := runtime.NumGoroutine() - 1 // subtract 1 for main thread
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, data.pieceIndex, numWorkers)
	}


	fmt.Println("###Total downloaded file size:", count)
	// Note: no need to manually close data_queue because the queue is ended when we receive all pieces(search text "##NOTE-1")
	// close(dataqueue)
	close(workqueue)
	return buf, nil
}

func (t *Torrent) spawnWorker(peer peer.Peer, workQueue chan *workmsg, dataqueue chan *datamsg) {
	client, err := client.New(peer, t.PeerId, t.InfoHash)
	if err != nil {
		log.Printf("Could not handshake with %s. Disconnecting\n", peer.IP)
		// wg.Done()
		return		
	}
	defer client.Conn.Close()
	fmt.Printf("Completed handshake with %s\n", peer.IP)

	err = client.SendUnchokeMsg()
	if err != nil {
		fmt.Println("###SendUnchokeMsg error:", err)
		// wg.Done()
		// return
	}
	err = client.SendInterestedMsg()
	if err != nil {
		fmt.Println("###SendInterestedMsg error:", err)
		// wg.Done()
		// return
	}

	fmt.Println("*******监听Workqueue")
	for msg := range workQueue {
		if !client.Bitfield.HasPiece(msg.pieceIndex) {
			workQueue <- msg // Put piece back on the queue
			continue
		}

		data, err := t.tryDownload(msg, client)
		if err != nil {
			workQueue <- msg //if something goes south, we put the work back to Queue
			continue
		}
		
		client.Bitfield.SetPiece(msg.pieceIndex)
		client.SendHaveMsg(msg.pieceIndex)
		
		

		dataqueue <- &datamsg {
			pieceIndex: msg.pieceIndex,
			data: data,
			
		}
			
	}

	// wg.Done()
} 




const maxPipelinedRequests int = 5
// MaxBlockSize is the largest number of bytes a request can ask for
const maxBlockSize = 16384

func (t *Torrent) tryDownload(msg *workmsg, c *client.Client) ([]byte, error) {
	/**
		Reference: https://blog.jse.li/posts/torrent/
	***/
	// Setting a deadline helps get unresponsive peers unstuck.
	// 30 seconds is more than enough time to download a 262 KB piece
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline

	// for each piece, we need to keep track of how many bytes we have downloaded
	// we use "state"
	state := stateProgress{
		index: msg.pieceIndex,
		c : c,
		buf: make([]byte, msg.lengthOfPiece),
	}
	
	
	for state.downloaded < msg.lengthOfPiece {
		if !state.c.Choked {
			for state.requestedByte < msg.lengthOfPiece && state.numOfPipelinedRequests < maxPipelinedRequests{
				
				
				blockSize := maxBlockSize
				// Why we are requesting by blocks but not by pieces?
				/**
					These two properties are necessary because sometimes a piece is too big for one message. Although there is some dispute about the best size, it is typically considered to be around 2^14 (16384) bytes. This is called a “block”, where a piece consists of one or more blocks. If the piece length is greater than the length of single block, then it should be broken up into blocks with one message sent for each block.

					So the “begin” field is the zero-based byte offset starting from the beginning of the piece, and the “length” is the length of the requested block. This is always going to be 2^14, except possibly the last block which might be less.

					reference: https://allenkim67.github.io/programming/2016/05/04/how-to-make-your-own-bittorrent-client.html#pieces-vs-blocks
				*/
				
				if blockSize > (msg.lengthOfPiece-state.requestedByte) {// there is a case where the last block length is less than the pieceLength
					blockSize = msg.lengthOfPiece - state.requestedByte
					
				}
				err := c.SendRequestMsg(msg.pieceIndex, state.requestedByte, blockSize)
				if err != nil {
					return nil, err
				}

				state.requestedBlock++
				state.requestedByte += blockSize
			}

		}


		err := state.ReadMessageFromPeer()
		if err != nil {
			return nil, err
		}
		
		
		fmt.Println("state.downloaded for pieceIndex:",msg.pieceIndex, " downloaded byte length:", len(state.buf))

	}

	return state.buf,nil
}



func (s *stateProgress) ReadMessageFromPeer() ( error) {
	message, err := msg.ReadChokeUnchokeMessage(s.c.Conn)
	if err != nil {

		return err
	}
	if message == nil { // keep-alive message
		return nil
	}

	switch message.MessageID {

		case msg.MsgIDUnchoke:
			s.c.Choked = false
		case msg.MsgIDPiece:
			fmt.Println("###Receive piece msg")
			numOfDownloadedBytes, errr := msg.ProcessPiece(s.index, s.buf, message)
			if errr != nil {
				return errr
			}
			s.downloaded += numOfDownloadedBytes
			s.numOfPipelinedRequests--
		case msg.MsgIDChoke:
			s.c.Choked = true
		// default:
		// 	fmt.Println("We are reading an unknown type of message from the peer")
		// 	return errors.New("We are reading an unknown type of message from the peer")
	}
	return  nil
}


func checkIntegrity(data []byte, expectedHash [20]byte) bool {
	

	h := sha1.Sum(data) // generate the SHA1 checksum of piece of data we receive from peers
	return bytes.Compare(h[:], expectedHash[:]) == 0
}