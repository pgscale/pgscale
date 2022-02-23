OUT=$(PWD)/bin
VERSION=0.1.0
ROOT_MODULE=github.com/buraksezer/pgscale-server
DESTDIR=$(shell go env GOPATH)/bin

GIT_SHA=`git rev-parse --short HEAD || echo "GitNotFound"`

build:
	go build -o ${OUT}/pgscale-server -ldflags="-s -w -X main.Version=${VERSION} -X main.GitSHA=${GIT_SHA}" ./cmd/pgscale-server

test:
	go test -v ./...

install:
	install ${OUT}/pgscale-server ${DESTDIR}

clean:
	rm -rf ${OUT}/pgscale-server

all: clean build install