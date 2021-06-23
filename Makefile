
VERSION = 0.1-1
TAG = $(VERSION)
PREFIX = nginx/nginx-ns1-gslb
TARGET ?= local
CONFIG_FILE=configs/example_global.yaml

all: nginx-ns1-gslb test lint container

nginx-ns1-gslb:
ifeq (${TARGET},local)
	$(eval GOPATH=$(shell go env GOPATH))
	CGO_ENABLED=0 GO111MODULE=on GOFLAGS="-gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH}" GOOS=linux go build -trimpath -ldflags "-s -w"  -o nginx-ns1-gslb github.com/nginxinc/nginx-ns1-gslb/cmd/agent
endif

test:
	GO111MODULE=on go test ./...

lint:
	docker run --pull always --rm -v $(shell pwd):/nginx-ns1-gslb -w /nginx-ns1-gslb -v $(shell go env GOCACHE):/cache/go -e GOCACHE=/cache/go -e GOLANGCI_LINT_CACHE=/cache/go -v $(shell go env GOPATH)/pkg:/go/pkg golangci/golangci-lint:latest golangci-lint --color always run


container: nginx-ns1-gslb
	docker build --build-arg CONFIG_FILE=$(CONFIG_FILE) -f build/Dockerfile --target $(TARGET) -t $(PREFIX):$(TAG) .

clean:
	rm -f nginx-ns1-gslb
