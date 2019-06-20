
#GOPATH := ${PWD}/../../../../
#export GOPATH

export GO111MODULE=auto

build:
	go build -o xbench cmd/main.go

test:
	go test test/*.go
  
lint:
	golint ./...

.PHONY: all test clean

