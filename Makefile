OUT=$(PWD)/bin
VERSION=0.1.0
ROOT_MODULE=github.com/buraksezer/dante-server
DESTDIR=$(shell go env GOPATH)/bin

GIT_SHA=`git rev-parse --short HEAD || echo "GitNotFound"`

build:
	go build -o ${OUT}/dante-server -ldflags="-s -w -X main.Version=${VERSION} -X main.GitSHA=${GIT_SHA}" ./cmd/dante-server

test:
	go test -v ./...

install:
	install ${OUT}/dante-server ${DESTDIR}

clean:
	rm -rf ${OUT}/dante-server

all: clean build install