# These are the values we want to pass for Version and BuildTime
GITTAG=`git describe --tags`

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X main.appVersion=${GITTAG}"

all: build

build: adapter forwarder

adapter:
	go build ${LDFLAGS} -o bin/httpdns-adapter ./cmd/adapter

forwarder:
	go build ${LDFLAGS} -o bin/httpdns-forwarder ./cmd/forwarder

.PHONY: adapter forwarder
