# Set the build version
ifeq ($(origin VERSION), undefined)
	VERSION := $(shell git rev-parse --short HEAD)
endif

# Setup some useful vars
HOST_GOOS = $(shell go env GOOS)
HOST_GOARCH = $(shell go env GOARCH)
GLIDE_VERSION = v0.11.1
ifeq ($(origin GLIDE_GOOS), undefined)
	GLIDE_GOOS := $(HOST_GOOS)
endif

travis-build: test
	GOOS=darwin go build -o smokeshift-darwin-amd64 -ldflags "-X main.version=$(VERSION)" ./cmd
	GOOS=linux go build -o smokeshift-linux-amd64 -ldflags "-X main.version=$(VERSION)" ./cmd

build: vendor
	go build -o bin/smokeshift -ldflags "-X main.version=$(VERSION)" ./cmd
	GOOS=darwin go build -o bin/smokeshift-darwin-amd64 -ldflags "-X main.version=$(VERSION)" ./cmd
	GOOS=linux go build -o bin/smokeshift-linux-amd64 -ldflags "-X main.version=$(VERSION)" ./cmd

clean:
	rm -rf bin
	rm -rf out
	rm -rf vendor

test: vendor
	go test ./cmd/... ./pkg/... $(TEST_OPTS)

vendor: tools/glide
	./tools/glide install

tools/glide:
	mkdir -p tools
	curl -L https://github.com/Masterminds/glide/releases/download/$(GLIDE_VERSION)/glide-$(GLIDE_VERSION)-$(GLIDE_GOOS)-$(HOST_GOARCH).tar.gz | tar -xz -C tools
	mv tools/$(GLIDE_GOOS)-$(HOST_GOARCH)/glide tools/glide
	rm -r tools/$(GLIDE_GOOS)-$(HOST_GOARCH)

