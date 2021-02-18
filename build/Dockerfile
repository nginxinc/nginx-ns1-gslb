FROM alpine:3.13.2
ARG CONFIG_FILE

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY nginx-ns1-gslb /
COPY ${CONFIG_FILE} /etc/nginx-ns1-gslb/config.yaml

ENTRYPOINT ["/nginx-ns1-gslb", "-config-file", "/etc/nginx-ns1-gslb/config.yaml"]