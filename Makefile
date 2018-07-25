# Absolute github repository name.
REPO := github.com/infobloxopen/atlas-app-toolkit

# Build directory absolute path.
PROJECT_ROOT = $(CURDIR)

# Utility docker image to generate Go files from .proto definition.
# https://github.com/infobloxopen/atlas-gentool
GENTOOL_IMAGE := infoblox/atlas-gentool:v3

.PHONY: default
default: test

test: check-fmt vendor
	go test -cover ./...

.PHONY: vendor
vendor:
	dep ensure

check-fmt:
	test -z `go fmt ./...`

.gen-query:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=:. $(REPO)/query/collection_operators.proto

.gen-errdetails:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=:. $(REPO)/rpc/errdetails/error_details.proto

.gen-errfields:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=:. $(REPO)/rpc/errfields/error_fields.proto

.gen-servertestdata:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=plugins=grpc:. --grpc-gateway_out=logtostderr=true:. $(REPO)/server/testdata/test.proto

.PHONY: gen
gen: .gen-op .gen-errdetails .gen-errfields
