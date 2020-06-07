package torrentfile

import(
	bencode "github.com/jackpal/bencode-go"
	"os"
	// "log"
	"fmt"
	"encoding/hex"
	
)

/**
	This module does the extracting info from torrent file.
**/

type TorrentFileRaw struct {
	Announce string 
	Info torrentMetaInfoRaw 
}

type TorrentFile struct {
	Announce string `bencode:"announce"`
	Info TorrentMetaInfo `bencode:"info"`
}

type TorrentMetaInfo struct {
	Length int `bencode:"length"`
	Name string `bencode:"name"`
	PieceLength int `bencode:"piece length"`
	Pieces [][20]byte `bencode:"pieces"`// each piece's sha1 hashed value
}

// Raw-version of the torrent file's "info" map
type torrentMetaInfoRaw struct {
	Length int 
	Name string 
	PieceLength int 
	Pieces string // each piece's sha1 hashed value
}

func Open(torrentFile string) *TorrentFile{
	file, err := os.Open("/mnt/d/bitTorrentClient/archlinux-2019.12.01-x86_64.iso.torrent")
	defer file.Close()
	checkErr(err)
	// fileinfo, err := file.Stat()
	// filesize := fileinfo.Size()
	// bufferReader := make([]byte, filesize)

	// bytesread, err :=file.Read(bufferReader)
	data := TorrentFileRaw{}
	err = bencode.Unmarshal(file, &data)
	checkErr(err)
	// log.Print("decoded bencode torrent file: ",data)
	TorrentMetaInfo := TorrentMetaInfo{data.Info.Length, data.Info.Name, data.Info.PieceLength, nil}
	TorrentMetaInfo.Pieces = *chopPieces(data.Info.Pieces, data.Info.PieceLength, data.Info.Length)
	// fmt.Println("tf obj created ", TorrentMetaInfo.Pieces)
	len := len(TorrentMetaInfo.Pieces)
	for i := 0; i < len; i++ {
		// encodedStr := hex.EncodeToString(chunksOfPieces[i])
		if i == len-1 {

			fmt.Printf("chunk %v %s\n", i,  hex.EncodeToString(TorrentMetaInfo.Pieces[i][:]))
		}
	}
	return  &TorrentFile{data.Announce, TorrentMetaInfo}
}

func chopPieces(pieces string, pieceLength int, length int) *[][20]byte {
	// Splits the hex string into a slice of byte arrays
	
	hashLen := 20
	piceseInOneByteArray := []byte(pieces)

	var numOfHahes int = len(piceseInOneByteArray) / hashLen
	fmt.Println(numOfHahes) // numOfHahes = 1278
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