# Absolute github repository name.
REPO := github.com/infobloxopen/atlas-app-toolkit

# Build directory absolute path.
PROJECT_ROOT = $(CURDIR)

# Utility docker image to build Go binaries
# https://github.com/infobloxopen/buildtool
BUILDTOOL_IMAGE := infoblox/buildtool:v8

# Utility docker image to generate Go files from .proto definition.
# https://github.com/infobloxopen/atlas-gentool
GENTOOL_IMAGE := infoblox/atlas-gentool:v3

BUILDER :=  docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) -w /go/src/$(REPO) $(BUILDTOOL_IMAGE)
# Set BUILD_TYPE environment variable to "local" in order to use local go instance instead of buildtool
ifeq ($(BUILD_TYPE), local)
BUILDER :=
endif

.PHONY: default
default: test

test: check-fmt vendor
	$(BUILDER) go test ./...

vendor:
	$(BUILDER) dep ensure

check-fmt:
	test -z `$(BUILDER) go fmt ./...`

.gen-op:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=:. $(REPO)/op/collection_operators.proto

.gen-errdetails:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=:. $(REPO)/rpc/errdetails/error_details.proto

.gen-servertestdata:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=plugins=grpc:. --grpc-gateway_out=logtostderr=true:. $(REPO)/server/testdata/test.proto

.PHONY: gen
gen: .gen-op .gen-errdetails
