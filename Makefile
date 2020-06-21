build:
	go build main.go
	rm test.txt
run:
	./main ubuntu-20.04-desktop-amd64.iso.torrent > test.txt
