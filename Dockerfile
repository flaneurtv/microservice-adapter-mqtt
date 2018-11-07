FROM golang:1.11 as builder
# Use this as a base to copy /usr/local/bin/microservice-adapter-mqtt from to
# be used in your multistage microservice builds.

ENV TARGET_ADAPTER_NAME microservice-adapter-mqtt
ENV TARGET_ADAPTER_PATH /usr/local/bin/$TARGET_ADAPTER_NAME

ENV TARGET_BRIDGE_NAME microservice-bridge-mqtt
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
ENV SERVICE_PROCESSOR=/srv/test-echo/processor
ENV SUBSCRIPTIONS=/srv/test-echo/subscriptions.txt

RUN apk add --no-cache bash jq gettext util-linux coreutils py2-pip && \
  pip install pytz

COPY --from=builder $TARGET_ADAPTER_PATH $TARGET_ADAPTER_PATH
COPY --from=builder $TARGET_BRIDGE_PATH $TARGET_BRIDGE_PATH

WORKDIR /srv/test-echo/
COPY . .

CMD ["microservice-adapter-mqtt"]
