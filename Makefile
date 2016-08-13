all: build

build:
	CGO_ENABLED=0 go build -ldflags '-s'

install:
	mv lf $(GOPATH)/bin

test:
	go test

.PHONY: all test
