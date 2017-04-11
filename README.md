# MQTT Service Playlist  Checker #

### docker run ###

FIXME: each service-processor will come with its own copy of the mqtt-service-adapter and will have its individual Dockerfile with its specific docker build instructions. The resulting docker image will carry the name of the specific service-processor run by the service-adapter: e.g. ticker-service or tick-responder-service. Therefore the line to start this thing will much rather look something like this:

For the ticker-service
```
docker rm -f ticker-service ; clear ; docker run -it --name ticker-service -v $HOME/www/ticker-service/:/usr/src/app/ -p 3000:3000 --link mqtt  -e NAMESPACE="flaneur" -e SERVICE_NAME="ticker-service" -e MQTT_LISTENER_URL="tcp://mqtt:1883" -e MQTT_PUBLISHER_URL="tcp://mqtt:1883" ticker-service
```

For the tick-responder-service
```
docker rm -f tick-responder-service ; clear ; docker run -it --name tick-responder-service -v $HOME/www/tick-responder-service/:/usr/src/app/ -p 3000:3000 --link mqtt  -e NAMESPACE="flaneur" -e SERVICE_NAME="tick-responder-service or taken from package.json" -e SERVICE_UUID="uuid now randomly assigned on start" -e SERVICE_HOST="docker id taken from the docker images hostname" -e MQTT_LISTENER_URL="tcp://mqtt:1883" -e MQTT_PUBLISHER_URL="tcp://mqtt:1883" -e MQTT_SUBSCRIPTIONS="flaneur/tick" tick-responder-service
```


TESTING docker
```
docker rm -f ticker-service ; clear ; docker run -it --name ticker-service -v $HOME/www/mqtt-service-adapter/service-adapter:/usr/src/app/service-adapter -v $HOME/www/mqtt-service-adapter/service-processor:/usr/src/app/service-processor -p 3000:3000 --link mqtt  -e NAMESPACE="flaneur" -e MQTT_LISTENER_URL="tcp://mqtt:1883" -e MQTT_PUBLISHER_URL="tcp://mqtt:1883" mqtt-service-adapter
```
