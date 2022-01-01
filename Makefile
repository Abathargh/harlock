harlock : cmd/harlock
	go build -ldflags "-s -w" ./cmd/harlock

install :
	go install -ldflags "-s -w" ./cmd/harlock

test :
	go test ./...

.PHONY : test