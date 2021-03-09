include Makefile.buf

# Absolute github repository name.
REPO := github.com/infobloxopen/atlas-app-toolkit

# Build directory absolute path.
PROJECT_ROOT = $(CURDIR)

# Utility docker image to generate Go files from .proto definition.
# https://github.com/infobloxopen/atlas-gentool
GENTOOL_IMAGE   := infoblox/docker-protobuf:latest
GENTOOL_OPTIONS := --rm -w /go/$(REPO) -v $(shell pwd):/go/$(REPO)
GENTOOL_FLAGS   := -I. \
	-Ithird-party \
	-Ivendor/github.com/grpc-ecosystem/grpc-gateway/v2 \
	--go_out=Mgoogle/protobuf/descriptor.proto=github.com/protocolbuffers/protobuf-go/types/descriptorpb,Mprotoc-gen-openapiv2/options/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options:/go
GATEWAY_FLAGS := --grpc-gateway_out=logtostderr=true:/go

GENERATOR := docker run $(GENTOOL_OPTIONS) $(GENTOOL_IMAGE) $(GENTOOL_FLAGS)


.PHONY: default
default: test

test: check-fmt lint vendor
	go test -cover ./...

lint: $(BUF)
	buf lint

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor

check-fmt:
	test -z `go fmt ./...`

query/collection_operators.pb.go: query/collection_operators.proto
	$(GENERATOR) \
		query/collection_operators.proto

atlas/atlasrpc/error_details.pb.go: proto/atlas/atlasrpc/v1/error_details.proto
	$(GENERATOR) $<

atlas/atlasrpc/error_fields.pb.go: proto/atlas/atlasrpc/v1/error_fields.proto
	$(GENERATOR) $<

atlas/resource/resource.pb.go: proto/atlas/resource/v1/resource.proto
	$(GENERATOR) $<

server/testdata/test.pb.go: server/testdata/test.proto
	$(GENERATOR) $(GATEWAY_FLAGS) $<

.PHONY: gen
gen: query/collection_operators.pb.go rpc/errdetails/error_details.pb.go rpc/errfields/error_fields.pb.go server/testdata/test.pb.go

bufgen: $(BUF)
	buf generate -o $(shell go env GOPATH)/src

.PHONY: mocks
mocks:
	GO111MODULE=off go get -u github.com/maxbrunsfeld/counterfeiter
	counterfeiter --fake-name ServerStreamMock -o ./logging/mocks/server_stream.go $(GOPATH)/src/github.com/infobloxopen/atlas-app-toolkit/vendor/google.golang.org/grpc/stream.go ServerStream
