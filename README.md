# MQTT Service Playlist  Checker #

### docker run ###
```
docker rm -f mqtt-service-adapter ; clear ; docker run -it --name mqtt-service-adapter -v $HOME/www/mqtt-service-adapter/deploy:/usr/src/app/deploy -p 3000:3000 --link mqtt  -e NAMESPACE="uuid" -e SERVICE_NAME="string" -e SERVICE_UUID="uuid" -e SERVICE_HOST="uuid" -e MQTT_LISTENER_URL="tcp://mqtt:1883" -e MQTT_PUBLISHER_URL="tcp://mqtt:1883" -e MQTT_SUBSCRIPTIONS="flaneur/tick;flaneur/tusd/upload_success" mqtt-service-adapter
```
```
