version := $(shell git describe --tags)
build :
	go build -ldflags "-s -w -X 'github.com/Abathargh/harlock/pkg/interpreter.Version=$(version)'" ./cmd/harlock

install :
	go install -ldflags "-s -w -X 'github.com/Abathargh/harlock/pkg/interpreter.Version=$(version)'" ./cmd/harlock

test :
	go test ./...

.PHONY : test
.PHONY : build
.PHONY : install
