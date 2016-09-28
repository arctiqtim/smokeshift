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

build: vendor
	go build "-X main.version=$(VERSION)" ./kuberang
	GOOS=linux go build -o bin/linux/$(HOST_GOARCH)/kuberang ./cmd/kuberang
	GOOS=darwin go build -o bin/darwin/$(HOST_GOARCH)/kuberang ./cmd/kuberang

clean:
	rm -rf bin
	rm -rf out

test: vendor
	go test ./cmd/... ./pkg/... $(TEST_OPTS)

integration-test: dist just-integration-test

vendor: tools/glide
	./tools/glide install

tools/glide:
	mkdir -p tools
	curl -L https://github.com/Masterminds/glide/releases/download/$(GLIDE_VERSION)/glide-$(GLIDE_VERSION)-$(GLIDE_GOOS)-$(HOST_GOARCH).tar.gz | tar -xz -C tools
	mv tools/$(GLIDE_GOOS)-$(HOST_GOARCH)/glide tools/glide
	rm -r tools/$(GLIDE_GOOS)-$(HOST_GOARCH)

vendor-ansible/out:
	docker build -t apprenda/vendor-ansible -q vendor-ansible
	docker run --rm -v $(shell pwd)/vendor-ansible/out:/ansible apprenda/vendor-ansible pip install --install-option="--prefix=/ansible" ansible

vendor-cfssl/out:
	mkdir -p vendor-cfssl/out/
	curl -L https://pkg.cfssl.org/R1.2/cfssl_linux-amd64 -o vendor-cfssl/out/cfssl_linux-amd64
	curl -L https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64 -o vendor-cfssl/out/cfssljson_linux-amd64
	curl -L https://pkg.cfssl.org/R1.2/cfssl_darwin-amd64 -o vendor-cfssl/out/cfssl_darwin-amd64
	curl -L https://pkg.cfssl.org/R1.2/cfssljson_darwin-amd64 -o vendor-cfssl/out/cfssljson_darwin-amd64

dist: vendor-ansible/out vendor-cfssl/out build
	mkdir -p out
	cp bin/kismatic out
	mkdir -p out/ansible
	cp -r vendor-ansible/out/* out/ansible
	rm -rf out/ansible/playbooks
	cp -r ansible out/ansible/playbooks
	mkdir -p out/ansible/playbooks/inspector
	cp -r bin/inspector/* out/ansible/playbooks/inspector
	mkdir -p out/cfssl
	cp -r vendor-cfssl/out/* out/cfssl
	rm -f out/kismatic.tar.gz
	tar -cvzf kismatic.tar.gz -C out .
	mv kismatic.tar.gz out