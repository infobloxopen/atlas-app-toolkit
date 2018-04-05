# Absolute github repository name.
REPO := github.com/Infoblox-CTO/ngp.api.toolkit
EXAMPLE := $(REPO)/example

# Build directory absolute path.
PROJECT_ROOT = $(CURDIR)
BINDIR = $(PROJECT_ROOT)/example/bin

# Utility docker image to build Go binaries
# https://github.com/infobloxopen/buildtool
BUILDTOOL_IMAGE := infoblox/buildtool:v8

# Utility docker image to generate Go files from .proto definition.
# https://github.com/Infoblox-CTO/ngp.api.toolkit/gentool
GENTOOL_IMAGE := infobloxcto/gentool:latest

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

.PHONY: gentool
gentool:
	cd gentool &&  docker build -f Dockerfile -t $(GENTOOL_IMAGE) .

.gen-op:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=:. $(REPO)/op/collection_operators.proto

.gen-errdetails:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=:. $(REPO)/rpc/errdetails/error_details.proto

.PHONY: gen
gen: .gen-op .gen-errdetails

.PHONY: example-gen
example-gen: example-gen-addressbook example-gen-dnsconfig

.PHONY: example-gen-addressbook
example-gen-addressbook: gentool
	 docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=plugins=grpc:. \
	--grpc-gateway_out=logtostderr=true:. \
	--validate_out="lang=go:." \
	$(EXAMPLE)/addressbook.proto


.PHONY: example-gen-dnsconfig
example-gen-dnsconfig: gentool
	 docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=plugins=grpc:. \
	--grpc-gateway_out=logtostderr=true:. \
	--validate_out="lang=go:." \
	$(EXAMPLE)/dnsconfig.proto

.PHONY: example-build
example-build: vendor
	$(BUILDER) go build -o "example/bin/gateway" "$(EXAMPLE)/cmd/gateway"
	$(BUILDER) go build -o "example/bin/addressbook" "$(EXAMPLE)/cmd/addressbook"
	$(BUILDER) go build -o "example/bin/dnsconfig" "$(EXAMPLE)/cmd/dnsconfig"

.PHONY: example-clean
example-clean:
	rm -rf "$(BINDIR)"

.PHONY: example-image
example-image:
	cd example &&  docker build -f Dockerfile.addressbook -t infobloxcto/addressbook:v1.0 .
	cd example &&  docker build -f Dockerfile.dnsconfig -t infobloxcto/dnsconfig:v1.0 .
	cd example &&  docker build -f Dockerfile.gateway -t infobloxcto/gateway:v1.0 .

.PHONY: example-image-clean
example-image-clean:
	docker rmi -f infobloxcto/addressbook:v1.0 infobloxcto/dnsconfig:v1.0 infobloxcto/gateway:v1.0

.PHONY: nginx-up
nginx-up:
	kubectl apply -f example/nginx.yaml

.PHONY: nginx-down
nginx-down:
	kubectl delete -f example/nginx.yaml

.PHONY: example-up
example-up:
	kubectl apply -f example/kube.yaml

.PHONY: example-down
example-down:
	kubectl delete -f example/kube.yaml
