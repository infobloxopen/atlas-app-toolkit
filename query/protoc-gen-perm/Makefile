GOPATH ?= $(HOME)/go
SRCPATH := $(patsubst %/,%,$(GOPATH))/src

default: options install


.PHONY: options
options:
	protoc -I. -I$(SRCPATH) -I../../vendor  \
		--gogo_out="Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:$(SRCPATH)" \
		options/collection_permissions.proto

.PHONY: install
install:
	go install

.PHONY: perm-test
perm-test: ./example/* ./test/*
	protoc -I. -I${SRCPATH} -I../../vendor 	--perm_out="./" example/example.proto
	go test  ./...
	