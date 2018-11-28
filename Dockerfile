FROM golang:1.11 as builder
# Use this as a base to copy /usr/local/bin/samm from to
# be used in your multistage microservice builds.

ENV TARGET_ADAPTER_NAME samm
ENV TARGET_ADAPTER_PATH /usr/local/bin/$TARGET_ADAPTER_NAME

ENV TARGET_BRIDGE_NAME sammbridge
ENV TARGET_BRIDGE_PATH /usr/local/bin/$TARGET_BRIDGE_NAME

COPY src /mqtt-adapter/src
COPY examples/. /srv/.

WORKDIR /mqtt-adapter/src

RUN go test -cover ./...
RUN GOOS=linux GOARCH=amd64 go build -ldflags "-linkmode external -extldflags -static" -o $TARGET_ADAPTER_PATH -a ./cmd/adapter
RUN GOOS=linux GOARCH=amd64 go build -ldflags "-linkmode external -extldflags -static" -o $TARGET_BRIDGE_PATH -a ./cmd/bridge


FROM alpine:3.8
# This demo service sends a tick message every 3 seconds. Use it as a blueprint
# for your own microservices using this adapter. It is recommended, to set ENV
# vars for everything specific to this service
# (SEVICE_NAME, SERVICE_PROCESSOR, SUBSCRIPTIONS) right here and infrastructure
# specific settings (MQTT_LISTENER_URL, auth credentials, ...) in
# docker-compose.yml and thelike.

ENV SERVICE_NAME=test-echo
ENV SERVICE_PROCESSOR=/srv/processor
ENV NAMESPACE_PUBLISHER=default
ENV NAMESPACE_LISTENER=default
ENV SUBSCRIPTIONS=/srv/subscriptions.txt

RUN apk add --no-cache bash jq gettext util-linux coreutils

COPY --from=builder /usr/local/bin/samm /usr/local/bin/samm
COPY --from=builder /usr/local/bin/sammbridge /usr/local/bin/sammbridge

WORKDIR /srv/
COPY examples/test-echo/. /srv/.

CMD ["samm"]
