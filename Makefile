version := $(git describe --tags)
harlock : cmd/harlock
	go build -ldflags "-s -w -X github.com/Abathargh/harlock/cmd/harlock.Version=$(version)" ./cmd/harlock

install :
	go install -ldflags "-s -w -X github.com/Abathargh/harlock/cmd/harlock.Version=$(version)" ./cmd/harlock

test :
	go test ./...

.PHONY : test