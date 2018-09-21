FROM golang:1.11-alpine3.8
# Use this as a base to copy /usr/local/bin/microservice-adapter-mqtt from to 
# be used in your multistage microservice builds.

ENV TARGET_NAME microservice-adapter-mqtt
ENV TARGET_PATH /usr/local/bin/$TARGET_NAME
ENV packages   ./adapter \
    			./config \
    			./logger \
    			./mqtt

RUN apk add --no-cache glide curl git make g++ \
  && curl -L https://git.io/vp6lP | sh \
  && apk add --no-cache bash jq gettext util-linux

RUN go get github.com/Masterminds/glide

COPY src /go/src/mqtt-adapter/src
COPY examples/. /srv/.

WORKDIR /go/src/mqtt-adapter/src

RUN glide install
RUN for package in $packages; do go test -cover -covermode=count $package; done
RUN GOOS=linux GOARCH=amd64 go build -o  $TARGET_PATH

CMD ["microservice-adapter-mqtt"]