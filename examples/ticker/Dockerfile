FROM flaneurtv/samm as adapter
# We will be copying /usr/local/bin/samm from this image.

FROM alpine:3.8
# This demo service sends a tick message every 3 seconds. Use it as a blueprint
# for your own microservices using this adapter. It is recommended, to set ENV
# vars for everything specific to this service
# (SEVICE_NAME, SERVICE_PROCESSOR, SUBSCRIPTIONS) right here and infrastructure
# specific settings (MQTT_LISTENER_URL, auth credentials, ...) in
# docker-compose.yml and thelike.

ENV SERVICE_NAME=ticker
ENV SERVICE_PROCESSOR=/srv/processor
ENV SUBSCRIPTIONS=/srv/subscriptions.txt

RUN apk add --no-cache bash jq gettext util-linux coreutils

COPY --from=adapter /usr/local/bin/samm /usr/local/bin/samm

WORKDIR /srv/
COPY . .

CMD ["samm"]
