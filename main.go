package main

import (
	torrentfile "kelvin94/btClient/torrentfile"
	"fmt"
	"net/http"
	"time"
)
type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}
func main()  {
	tf := torrentfile.Open("")
	url, err := tf.BuildTrackerGETRequest()
	if err != nil {
		fmt.Println( err)
	}

	c := &http.Client{Timeout: 30 * time.Second}
	res, err := c.Get(url)
	if err != nil {
		fmt.Println( err)
	}
	fmt.Println(res)
	// fmt.Println()
	// get TorrentFile Object

	// dump TorrentFile into Tracker
}