FROM golang:1.11
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

FROM scratch

COPY --from=0 $TARGET_ADAPTER_PATH $TARGET_ADAPTER_PATH
COPY --from=0 $TARGET_BRIDGE_PATH $TARGET_BRIDGE_PATH

CMD ["microservice-adapter-mqtt"]
