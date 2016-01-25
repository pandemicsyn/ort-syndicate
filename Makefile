SHA := $(shell git rev-parse --short HEAD)
VERSION := $(shell cat VERSION)
ITTERATION := $(shell date +%s)

deps:
	go get -u google.golang.org/grpc
	go get -u github.com/golang/protobuf/proto
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u github.com/gholt/ring
	go get -u github.com/gholt/ring/ring
	go get -u github.com/gholt/store
	go get -u github.com/pandemicsyn/cmdctrl
	go get -u github.com/pandemicsyn/cmdctrl/api

build:
	mkdir -p packaging/output
	mkdir -p packaging/root/usr/local/bin
	go build -i -v -o packaging/root/usr/local/bin/synd --ldflags " \
		-X main.ringVersion=$(shell git -C $$GOPATH/src/github.com/gholt/ring rev-parse HEAD) \
		-X main.syndVersion=$(shell git rev-parse HEAD) \
		-X main.goVersion=$(shell go version | sed -e 's/ /-/g') \
		-X main.buildDate=$(shell date -u +%Y-%m-%d.%H:%M:%S)" github.com/pandemicsyn/syndicate/synd 
	go build -i -v -o packaging/root/usr/local/bin/syndicate-client --ldflags " \
		-X main.ringVersion=$(shell git -C $$GOPATH/src/github.com/gholt/ring rev-parse HEAD) \
		-X main.syndicateClientVersion=$(shell git rev-parse HEAD) \
		-X main.goVersion=$(shell go version | sed -e 's/ /-/g') \
		-X main.buildDate=$(shell date -u +%Y-%m-%d.%H:%M:%S)"  github.com/pandemicsyn/syndicate/syndicate-client

clean:
	rm -rf packaging/output
	rm -f packaging/root/usr/local/bin/synd
	rm -f packaging/root/usr/local/bin/syndicate-client

install:
	#install -t /usr/local/bin packaging/root/usr/local/bin/synd
	go install --ldflags " \
		-X main.ringVersion=$(RINGVERSION) \
		-X main.syndVersion=$(shell git rev-parse HEAD) \
		-X main.goVersion=$(shell go version | sed -e 's/ /-/g') \
		-X main.buildDate=$(shell date -u +%Y-%m-%d.%H:%M:%S)" github.com/pandemicsyn/syndicate/synd 
	go install --ldflags " \
		-X main.ringVersion=$(shell git -C $$GOPATH/src/github.com/gholt/ring rev-parse HEAD) \
		-X main.syndicateClientVersion=$(shell git rev-parse HEAD) \
		-X main.goVersion=$(shell go version | sed -e 's/ /-/g') \
		-X main.buildDate=$(shell date -u +%Y-%m-%d.%H:%M:%S)"  github.com/pandemicsyn/syndicate/syndicate-client

run:
	go run synd/*.go

ring:
	go get github.com/gholt/ring/ring
	go install github.com/gholt/ring/ring
	mkdir -p /etc/oort/ring/value
	mkdir -p /etc/oort/ring/group
	ring /etc/oort/ring/value/valuestore.builder create replicas=3
	ring /etc/oort/ring/value/valuestore.builder add active=true capacity=1000 tier0=removeme 
	ring /etc/oort/ring/value/valuestore.builder ring
	ring /etc/oort/ring/group/groupstore.builder create replicas=3
	ring /etc/oort/ring/group/groupstore.builder add active=true capacity=1000 tier0=removeme 
	ring /etc/oort/ring/group/groupstore.builder ring
