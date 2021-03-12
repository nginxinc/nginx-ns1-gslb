all: container

VERSION = 0.1-1
TAG = $(VERSION)
PREFIX = nginx/nginx-ns1-gslb

DOCKER_RUN = docker run --rm -v $(shell pwd):/go/src/github.com/nginxinc/nginx-ns1-gslb
DOCKER_BUILD_RUN = docker run --rm -v $(shell pwd):/go/src/github.com/nginxinc/nginx-ns1-gslb -w /go/src/github.com/nginxinc/nginx-ns1-gslb/cmd/agent/
BUILD_IN_CONTAINER = 1
DOCKERFILEPATH = build
GOLANG_CONTAINER = golang:1.16
CONFIG_FILE=configs/example_global.yaml

nginx-ns1-gslb:
ifeq ($(BUILD_IN_CONTAINER),1)
	$(DOCKER_BUILD_RUN) -e CGO_ENABLED=0 $(GOLANG_CONTAINER) go build -installsuffix cgo -ldflags "-w" -o /go/src/github.com/nginxinc/nginx-ns1-gslb/nginx-ns1-gslb
else
	CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -ldflags "-w" -o nginx-ns1-gslb github.com/nginxinc/nginx-ns1-gslb/cmd/agent
endif

test:
ifeq ($(BUILD_IN_CONTAINER),1)
	$(DOCKER_RUN) $(GOLANG_CONTAINER) go test ./...
else
	go test ./...
endif

lint:
	golangci-lint run

container: test nginx-ns1-gslb
	docker build --build-arg CONFIG_FILE=$(CONFIG_FILE) -f $(DOCKERFILEPATH)/Dockerfile -t $(PREFIX):$(TAG) .

clean:
	rm -f nginx-ns1-gslb
	rm -f Dockerfile
