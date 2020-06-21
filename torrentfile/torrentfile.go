package torrentfile

import(
	bencode "github.com/jackpal/bencode-go"
	"os"
	"bytes"
	// "log"
	// "fmt"
	// "encoding/hex"
	"crypto/sha1"
)

/**
	This module does the extracting info from torrent file.
**/

type bencodeTorrent struct {
	Announce string `bencode:"announce"`
	Info TorrentInfoRaw  `bencode:"info"`
}

type TorrentFile struct {
	InfoHash    [20]byte
	Announce string `bencode:"announce"`
	Info TorrentInfo `bencode:"info"`	
}

type TorrentInfo struct {
	
	Length int `bencode:"length"`
	Name string `bencode:"name"`
	PieceLength int `bencode:"piece length"`
	Pieces [][20]byte `bencode:"pieces"`// each data_piece's sha1 hashed value
}

// Raw-version of the torrent file's "info" map
type TorrentInfoRaw struct {
	Length int  `bencode:"length"`
	Name string `bencode:"name"`
	PieceLength int `bencode:"piece length"`
	Pieces string `bencode:"pieces"` // each piece's sha1 hashed value
}


/**
	generateInfoHash(): Apply sha1 encryption into the entire "Info" section of the torrent file
	Return: [20]byte 
***/
func (tmi *TorrentInfoRaw)generateInfoHash() ([20]byte,error) {
	var b bytes.Buffer // bytes.Buffer struct implements io.Writer interface as well!!!
	err := bencode.Marshal(&b,*tmi)	
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(b.Bytes()) // generate the SHA1 checksum of the torrentfile.info
	return h,nil
}

func Open(torrentFilePath string) *TorrentFile{
	file, err := os.Open(torrentFilePath)
	defer file.Close()
	checkErr(err)
	// fileinfo, err := file.Stat()
	// filesize := fileinfo.Size()
	// bufferReader := make([]byte, filesize)

	// bytesread, err :=file.Read(bufferReader)
	data := bencodeTorrent{}
	err = bencode.Unmarshal(file, &data)
	checkErr(err)
	TorrentInfo := TorrentInfo{data.Info.Length, data.Info.Name, data.Info.PieceLength, nil}
	
	infoHash,err := data.Info.generateInfoHash()
	checkErr(err)
	temp := chopPieces(data.Info.Pieces, data.Info.PieceLength, data.Info.Length)	
	TorrentInfo.Pieces = *temp	
	
	return  &TorrentFile{infoHash,data.Announce, TorrentInfo}
}

func chopPieces(pieces string, pieceLength int, length int) *[][20]byte {
	// Splits the hex string into a slice of byte arrays
	
	hashLen := 20
	piceseInOneByteArray := []byte(pieces)

	var numOfHahes int = len(piceseInOneByteArray) / hashLen
	// Why one byte = one hex string character?
	// Each hexadecimal digit represents 4 binary digits, also known as a nibble, which is half a byte. For example, a single byte can have values ranging from 00000000 to 11111111 in binary form, which can be conveniently represented as 00 to FF in hexadecimal.

	var chunksOfPieces [][20]byte 
	chunksOfPieces = make([][20]byte,numOfHahes) 
	for i := 0; i < numOfHahes; i++ {
		copy(chunksOfPieces[i][:], piceseInOneByteArray[i*20:(i+1)*20])
	}

	// debug use only
	// fmt.Printf("%s\n", "Printing pieces to chunks of Hex string before")
	// for i := 0; i < numOfHahes; i++ {
	// 	if i == numOfHahes-1 {
	// 		fmt.Printf("chunk %v %s\n", i,  hex.EncodeToString(chunksOfPieces[i][:]))
	// 	}
	// }
	return &chunksOfPieces
}


func checkErr(e error) {
    if e != nil {
        panic(e)
    }
}