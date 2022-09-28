version := $(shell git describe --tags)
build :
	go build -ldflags "-s -w -X 'github.com/Abathargh/harlock/pkg/interpreter.Version=$(version)'" ./cmd/harlock

install :
	go install -ldflags "-s -w -X 'github.com/Abathargh/harlock/pkg/interpreter.Version=$(version)'" ./cmd/harlock

test :
	go test ./...

dist :
	make deb
	make standalone

deb :
	bash ./scripts/build-deb.sh

standalone :
	bash ./scripts/build-standalone.sh

.PHONY : deb
.PHONY : dist
.PHONY : standalone
.PHONY : test
.PHONY : build
.PHONY : build-interrepl
.PHONY : install
