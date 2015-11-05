.PHONY: all install test vet

all: test

install:
	go install

test:
	go test

vet:
	go vet
	golint .
