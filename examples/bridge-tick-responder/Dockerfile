FROM flaneurtv/samm as adapter
# We will be copying /usr/local/bin/samm from this image.

FROM alpine:3.8
# This demo service sends a tick message every 3 seconds. Use it as a blueprint
# for your own microservices using this adapter. It is recommended, to set ENV
# vars for everything specific to this service
# (SEVICE_NAME, SERVICE_PROCESSOR, SUBSCRIPTIONS) right here and infrastructure
# specific settings (MQTT_LISTENER_URL, auth credentials, ...) in
# docker-compose.yml and thelike.

ENV SERVICE_NAME=bridge-tick-responder
ENV SUBSCRIPTIONS=/srv/subscriptions.txt

COPY --from=adapter /usr/local/bin/sammbridge /usr/local/bin/sammbridge

WORKDIR /srv/
COPY . .

CMD ["sammbridge"]
