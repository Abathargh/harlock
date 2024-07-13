version := $(shell git describe --tags)

all : build

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

clean :
	rm -rf dist harlock


.PHONY : deb
.PHONY : dist
.PHONY : standalone
.PHONY : test
.PHONY : build
.PHONY : build-interrepl
.PHONY : install
.PHONY : clean