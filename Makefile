# Absolute github repository name.
REPO := github.com/infobloxopen/atlas-app-toolkit

# Build directory absolute path.
PROJECT_ROOT = $(CURDIR)

# Utility docker image to generate Go files from .proto definition.
# https://github.com/infobloxopen/atlas-gentool
GENTOOL_IMAGE := infoblox/atlas-gentool:latest

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

.gen-query:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=:. $(REPO)/query/collection_operators.proto
	mv $(PROJECT_ROOT)/v2/query/collection_operators.pb.go $(PROJECT_ROOT)/query/collection_operators.pb.go
	rm -rf $(PROJECT_ROOT)/v2

.gen-errdetails:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=:. $(REPO)/rpc/errdetails/error_details.proto
	mv $(PROJECT_ROOT)/rpc/errdetails/error_details.pb.go $(PROJECT_ROOT)/rpc/errdetails/error_details.pb.go
	rm -rf $(PROJECT_ROOT)/v2

.gen-errfields:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=:. $(REPO)/rpc/errfields/error_fields.proto
	mv $(PROJECT_ROOT)/rpc/errfields/error_fields.pb.go $(PROJECT_ROOT)/rpc/errfields/error_fields.pb.go
	rm -rf $(PROJECT_ROOT)/v2

.gen-servertestdata:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=plugins=grpc:. --grpc-gateway_out=logtostderr=true:. $(REPO)/server/testdata/test.proto

.PHONY: gen
gen: .gen-query .gen-errdetails .gen-errfields

.PHONY: mocks
mocks:
	GO111MODULE=off go get -u github.com/maxbrunsfeld/counterfeiter
	counterfeiter --fake-name ServerStreamMock -o ./logging/mocks/server_stream.go $(GOPATH)/src/github.com/infobloxopen/atlas-app-toolkit/vendor/google.golang.org/grpc/stream.go ServerStream
