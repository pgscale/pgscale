VERSION=0.1.0
ROOT_MODULE=github.com/buraksezer/pgscale-server
DESTDIR=$(shell go env GOPATH)/bin

GIT_SHA=`git rev-parse --short HEAD || echo "GitNotFound"`

build:
	go build -o ${DESTDIR}/pgscale-server -ldflags="-s -w -X main.Version=${VERSION} -X main.GitSHA=${GIT_SHA}" ./cmd/pgscale-server

test:
	go test -v ./...

clean:
	if [ -d ${DESTDIR}/pgscale-server ]; then rm -rf ${DESTDIR}/pgscale-server; fi

all: clean build
