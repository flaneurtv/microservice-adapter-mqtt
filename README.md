# MQTT Service Adapter #

Each service-processor comes with its own copy of the mqtt-service-adapter repository as a git submodule under ./service-adapter dir and will have its individual Dockerfile with its specific docker build instructions. The resulting docker image will carry the name of the specific service-processor run by the service-adapter: e.g. micro-ticker or micro-tick-responder.
