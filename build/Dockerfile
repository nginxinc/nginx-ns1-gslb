FROM alpine:3.13 as base
ARG CONFIG_FILE

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY ${CONFIG_FILE} /etc/nginx-ns1-gslb/config.yaml

ENTRYPOINT ["/nginx-ns1-gslb", "-config-file", "/etc/nginx-ns1-gslb/config.yaml"]


FROM golang:1.16-alpine as builder
WORKDIR /go/src/github.com/nginxinc/nginx-ns1-gslb/cmd/agent
COPY go.mod go.sum /go/src/github.com/nginxinc/nginx-ns1-gslb/
RUN go mod download
COPY . /go/src/github.com/nginxinc/nginx-ns1-gslb/
RUN GO111MODULE=on CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /nginx-ns1-gslb


FROM base AS container
COPY --from=builder /nginx-ns1-gslb /


FROM base AS local
COPY nginx-ns1-gslb /
