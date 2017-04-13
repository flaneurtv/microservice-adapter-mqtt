# MQTT Service Playlist  Checker #

### docker run ###

FIXME: each service-processor will come with its own copy of the mqtt-service-adapter and will have its individual Dockerfile with its specific docker build instructions. The resulting docker image will carry the name of the specific service-processor run by the service-adapter: e.g. ticker-service or tick-responder-service. Therefore the line to start this thing will much rather look something like this:

Building the docker image
```
docker build -t flaneurtv/ticker-service --no-cache .
```

Testing the service-processors directly
```
docker rm -f sandbox ; clear ; docker run -it --name sandbox -v $PWD/service-adapter:/usr/src/app/service-adapter -v $PWD/service-processor:/usr/src/app/service-processor --link mqtt -e NAMESPACE="flaneur" -e SERVICE_NAME="ticker-service" -e MQTT_LISTENER_URL="tcp://mqtt:1883" -e MQTT_PUBLISHER_URL="tcp://mqtt:1883" flaneurtv/ticker-service /bin/bash
```

Then run the following commands inside the container and see what happens
```
./service-processor/processor
./service-processor/processor | ./service-processor/processor-tick-responder
```

For the ticker-service
```
docker rm -f ticker-service ; clear ; docker run -it --name ticker-service -v $PWD/service-adapter:/usr/src/app/service-adapter -v $PWD/service-processor:/usr/src/app/service-processor --link mqtt -e NAMESPACE="flaneur" -e SERVICE_NAME="ticker-service" -e MQTT_LISTENER_URL="tcp://mqtt:1883" -e MQTT_PUBLISHER_URL="tcp://mqtt:1883" flaneurtv/ticker-service
```

For the tick-responder-service
```
docker rm -f tick-responder-service ; clear ; docker run -it --name tick-responder-service -v $PWD/service-adapter:/usr/src/app/service-adapter -v $PWD/service-processor:/usr/src/app/service-processor --link mqtt -e NAMESPACE="flaneur" -e SERVICE_NAME="tick-responder-service" -e MQTT_LISTENER_URL="tcp://mqtt:1883" -e MQTT_PUBLISHER_URL="tcp://mqtt:1883" -e SERVICE_PROCESSOR="./service-processor/processor-tick-responder" flaneurtv/ticker-service
```
