FROM alpine:3.18 as base
ARG CONFIG_FILE

RUN apk add --no-cache ca-certificates

COPY ${CONFIG_FILE} /etc/nginx-ns1-gslb/config.yaml

ENTRYPOINT ["/nginx-ns1-gslb", "-config-file", "/etc/nginx-ns1-gslb/config.yaml"]


FROM golang:1.19.0-alpine as builder
WORKDIR /go/src/github.com/nginxinc/nginx-ns1-gslb/cmd/agent
COPY go.mod go.sum /go/src/github.com/nginxinc/nginx-ns1-gslb/
RUN go mod download
COPY . /go/src/github.com/nginxinc/nginx-ns1-gslb/
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /nginx-ns1-gslb


FROM base AS container
COPY --from=builder /nginx-ns1-gslb /


FROM base AS local
COPY nginx-ns1-gslb /

FROM base AS goreleaser
ARG TARGETARCH
ARG TARGETPLATFORM

LABEL org.nginx.ns1-gslb.image.build.target="${TARGETPLATFORM}"
LABEL org.nginx.ns1-gslb.image.build.version="goreleaser"

COPY dist/nginx-ns1-gslb_linux_$TARGETARCH*/nginx-ns1-gslb /
