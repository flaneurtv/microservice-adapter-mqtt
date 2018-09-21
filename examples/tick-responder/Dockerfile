FROM flaneurtv/microservice-adapter-mqtt as adapter
# We will be copying /usr/local/bin/microservice-adapter-mqtt from this image.

FROM golang:1.11-alpine3.8
# This demo service sends a tick message every 3 seconds. Use it as a blueprint
# for your own microservices using this adapter. It is recommended, to set ENV 
# vars for everything specific to this service 
# (SEVICE_NAME, SERVICE_PROCESSOR, SUBSCRIPTIONS) right here and infrastructure 
# specific settings (MQTT_LISTENER_URL, auth credentials, ...) in 
# docker-compose.yml and thelike.

ENV SERVICE_NAME=tick-responder
ENV SERVICE_PROCESSOR=/srv/tick-responder/processor-tick-responder
# SUBSCRIPTIONS is not yet implemented
ENV SUBSCRIPTIONS=/srv/tick-responder/subscriptions.txt

RUN apk add --no-cache bash jq gettext util-linux coreutils

COPY --from=adapter /usr/local/bin/microservice-adapter-mqtt /usr/local/bin/microservice-adapter-mqtt

WORKDIR /srv/tick-responder/
COPY . .

CMD ["microservice-adapter-mqtt","--subs=/srv/tick-responder/subscriptions.txt"]