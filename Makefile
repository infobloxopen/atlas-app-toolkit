# Absolute github repository name.
REPO := github.com/infobloxopen/atlas-app-toolkit

# Build directory absolute path.
PROJECT_ROOT = $(CURDIR)

# Utility docker image to generate Go files from .proto definition.
# https://github.com/infobloxopen/atlas-gentool
GENTOOL_IMAGE := infoblox/atlas-gentool:v2

default: test

test: check-fmt vendor
	echo "" > coverage.txt
	for d in `go list ./... | grep -v vendor`; do \
		t=$$(date +%s); \
		go test -v -coverprofile=cover.out -covermode=atomic $$d || exit 1; \
		echo "Coverage test $$d took $$(($$(date +%s)-t)) seconds"; \
		if [ -f cover.out ]; then \
			cat cover.out >> coverage.txt; \
			rm cover.out; \
		fi; \
    done

.PHONY: vendor
vendor:
	dep ensure

check-fmt:
	go fmt ./...

.gen-op:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=:. $(REPO)/op/collection_operators.proto

.gen-errdetails:
	docker run --rm -v $(PROJECT_ROOT):/go/src/$(REPO) $(GENTOOL_IMAGE) \
	--go_out=:. $(REPO)/rpc/errdetails/error_details.proto

gen: .gen-op .gen-errdetails
