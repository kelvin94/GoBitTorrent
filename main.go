package main

import (
	torrentfile "kelvin94/btClient/torrentfile"

	"fmt"
	"os"
	downloader "kelvin94/btClient/downloader"

)


func main()  {
	torrentFilePath := os.Args[1]
	tf := torrentfile.Open(torrentFilePath)
	
	url, peerId, err := tf.BuildTrackerGETRequest()
	if err != nil {
		fmt.Println( err)
	}
	peers,err := tf.CallTracker(url)
	if err != nil {
		fmt.Println(err)
	}
	torrent := &downloader.Torrent{
		Peers : peers,
		InfoHash : tf.InfoHash,
		PeerId : peerId, // my own peer ID
		PieceLength : tf.Info.PieceLength,
		Pieces  : tf.Info.Pieces,
		Length  : tf.Info.Length,
	}


	file, err := torrent.Download()
	if(len(file) > 0) {
		fmt.Println("###Finished download, the file size is:",len(file))
	}
	
	if err != nil {
		fmt.Println(err)
	}

	outFile, err := os.Create("/mnt/d/bitTorrentClient/debian.iso")
	if err != nil {
		fmt.Println(err) 
	}
	defer outFile.Close()
	_, err = outFile.Write(file)
	if err != nil {
		fmt.Println(err)
	}


}