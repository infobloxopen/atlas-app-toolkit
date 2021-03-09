# Absolute github repository name.
REPO := github.com/infobloxopen/atlas-app-toolkit

# Build directory absolute path.
PROJECT_ROOT = $(CURDIR)

# Utility docker image to generate Go files from .proto definition.
# https://github.com/infobloxopen/atlas-gentool
GENTOOL_IMAGE   := infoblox/docker-protobuf:latest
GENTOOL_OPTIONS := --rm -w /go/$(REPO) -v $(shell pwd):/go/$(REPO)
GENTOOL_FLAGS   := -I. -Ivendor \
	-Ivendor/github.com/grpc-ecosystem/grpc-gateway/v2 \
	--go_out=Mgoogle/protobuf/descriptor.proto=github.com/protocolbuffers/protobuf-go/types/descriptorpb,Mprotoc-gen-openapiv2/options/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options:/go
GATEWAY_FLAGS := --grpc-gateway_out=logtostderr=true:/go

GENERATOR := docker run $(GENTOOL_OPTIONS) $(GENTOOL_IMAGE) $(GENTOOL_FLAGS)


.PHONY: default
default: test

test: check-fmt vendor
	go test -cover ./...

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor

check-fmt:
	test -z `go fmt ./...`

query/collection_operators.pb.go: query/collection_operators.proto
	$(GENERATOR) \
		query/collection_operators.proto

rpc/errdetails/error_details.pb.go: rpc/errdetails/error_details.proto
	$(GENERATOR) \
		rpc/errdetails/error_details.proto

rpc/errfields/error_fields.pb.go: rpc/errfields/error_fields.proto
	$(GENERATOR) \
		rpc/errfields/error_fields.proto

rpc/resource/resource.pb.go: rpc/resource/resource.proto
	$(GENERATOR) \
		rpc/resource/resource.proto

server/testdata/test.pb.go: server/testdata/test.proto
	$(GENERATOR) $(GATEWAY_FLAGS) \
		server/testdata/test.proto

.PHONY: gen
gen: query/collection_operators.pb.go rpc/errdetails/error_details.pb.go rpc/errfields/error_fields.pb.go server/testdata/test.pb.go

.PHONY: mocks
mocks:
	GO111MODULE=off go get -u github.com/maxbrunsfeld/counterfeiter
	counterfeiter --fake-name ServerStreamMock -o ./logging/mocks/server_stream.go $(GOPATH)/src/github.com/infobloxopen/atlas-app-toolkit/vendor/google.golang.org/grpc/stream.go ServerStream
