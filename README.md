# MQTT Service Playlist  Checker #

### docker run ###
```
docker rm -f tusd-hooks ; clear ; docker run -it --name tusd-hooks -v $HOME/tusd-upload:/srv/tusd-data -v $HOME/www/tusd-hooks-docker/local:/run/secrets -p 1080:1080 -v $PWD/hooks:/srv/tusd-hooks --link mqtt flaneurtv/tusd-hooks
```
