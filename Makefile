
#GOPATH := ${PWD}/../../../../
#export GOPATH

export GO111MODULE=auto

OBJDIR := output

build:
	mkdir -p $(OBJDIR)
	go build -o $(OBJDIR)/bin/xbench cmd/main.go
	go build --buildmode=plugin -o $(OBJDIR)/plugins/crypto/crypto-default.so.1.0.0 github.com/xuperchain/xuperunion/crypto/client/xchain/plugin_impl
	go build --buildmode=plugin -o $(OBJDIR)/plugins/crypto/crypto-schnorr.so.1.0.0 github.com/xuperchain/xuperunion/crypto/client/schnorr/plugin_impl
	cp -r conf $(OBJDIR)
	cp -r data $(OBJDIR)

test:
	go test test/*.go
  
lint:
	golint ./...

clean:
	rm -rf $(OBJDIR)

.PHONY: all test clean

