# SAMM - Service Adapter for MQTT Microservices #

SAMM is a building block for creating MQTT enabled microservices. SAMM connects to an MQTT message bus, subscribes to your topics and forwards the received messages to your custom processors stdin in JSON - one message == one line. If your processor wants to emit a message, you only need to write a JSON line to stdout.

##### Benefits #####
* Separation of concern: SAMM for MQTT boiler-plate, your code for functionality
* This is NOT a library or module.
* Write service processors in any language or even in BASH.
* SAMM comes in binary, making integration with your Docker runtime trivial.
* Integrated logging to MQTT or stderr.
* Effortless upgrades or moving to other messages bus protocol.
* Convention over configuration paradigm.
* MIT Open Source license.

### Overview of SAMM Architecture ###
The SAMM architecture implements a separation of concern:
* Communication: SAMM - Golang Service Adapter MQTT
* Functionality: we call this the Service Processor (your code)

SAMM will execute your service processor as a continuously running subprocess establishing communication to your processor via its stdin/out/err pipes.
```
  |  +-------------------------------------------------+
 M|  |    samm                   starts subprocess     |
 Q|  |                        +---------------------+  |
 T|  |                        |      processor      |  |
 T|  |                        |                     |  |
  +-----> subscribe        ------> stdin            |  |
  +-----< publish          <------ stdout           |  |
 B|  |                     <------ stderr           |  |
 U|  |                        +---------------------+  |
 S|  +-------------------------------------------------+
```
[^binaries]

##### Functionality in Detail #####
* SAMM is parametrized at start-up with a list of topics for subscription.
* MQTT messages received by SAMM will be forwarded to your processor, which will receive them on its stdin.
* All communication with processors is performed in the form of JSON one-liners.
* Each line equals a single message.
* Your processor can respond with JSON messages written to stdout, which will be published to MQTT by SAMM.
* These messages therefore have one required field {"topic":"string"}.
* Messages to stderr are treated as log messages by SAMM and can have a loglevel assigned.
* There is a specialised error JSON schema. Anything written to stderr, which is not formatted in JSON is treated as loglevel error.
* If your processor dies, SAMM will exit as well.

##### Bridge Mode #####
SAMM comes with an extra binary for bridge mode - sammbridge - which allows for easy bridging of subscribed messages from MQTT_LISTENER_URL to MQTT_PUBLISHER_URL.

### Docker image ###

The Dockerfile serves as an example on how to make use of this image in your multistage microservice builds. Find the Docker image here: https://hub.docker.com/r/flaneurtv/samm/

```
docker pull flaneurtv/samm
```

### Configuration, Schemata and Environment ###
* SERVICE_NAME (should be provided by you)
* SERVICE_HOST (usually determined by hostname call)
* SERVICE_UUID (usually omitted as random UUID assigned if none provided)
* SUBSCRIPTIONS (default is /srv/subscriptions.txt)
* NAMESPACE_LISTENER ("default" if unset)
* NAMESPACE_PUBLISHER ("default" if unset)
* NAMESPACE a convenience variable it the above tow are equal. Sets NAMESPACE_LISTENER and NAMESPACE_PUBLISHER and exposes them to service processor. The NAMESPACE variable is NOT exposed to the processor.
* MQTT_LISTENER_URL (default is "tcp://mqtt:1883")
* MQTT_PUBLISHER_URL (default is "tcp://mqtt:1883")
* MQTT_LISTENER_CREDENTIALS (default is /run/secrets/mqtt_listener.json)
* MQTT_PUBLISHER_CREDENTIALS (default is /run/secrets/mqtt_publisher.json)
* LOG_LEVEL (default is "error"; one of [debug|info|notice|warning|error|critical|alert|emergency])
* LOG_LEVEL_MQTT (default is "error"; same available as above)

##### Message Schema #####
The message schema should look like that but is not enforced. The only required field is "topic", which MUST NOT contain these characters: # +

We recommend to make your service as generic as possible, which also applies to the topic you choose for message publishing. Make sure to prepend the generic topic with $NAMESPACE_PUBLISHER env variable, to allow for personalisation. $NAMESPACE_LISTENER is used when SAMM subscribes to messages. There is two, so you can receive messages from Bus A but publish results to Bus B. Could also be used to build a message bridge, but we provide a special executable for this very purpose.
```
{
  "topic": "$NAMESPACE_PUBLISHER/$YOURTOPIC",
  "service_uuid": "$SERVICE_UUID",
  "service_name": "$SERVICE_NAME",
  "service_host": "$SERVICE_HOST",
  "created_at": "$CREATED_AT",
  "payload": {
  }
}
```

##### Error Messages #####
```
{
  "log_level": "debug",
  "log_message": "$TICK_UUID $TICK_TIMESTAMP",
  "created_at": "$CREATED_AT"
}
```
\* Contrary to this layout, message must be formatted as a one-liner terminated by a newline but without newlines in between.

##### MQTT Credentials #####
```
{
  "username": "username",
  "password": "password"
}
```

### Examples ###

A few examples are contained in this repository. The Dockerfile by default creates a container spinning up an echo-sender service, which - when hooked up to an MQTT broker - will echo everything you send to topic "default/test" back on topic "default/log/test".

A more complex and integrated examples can be spun up using the docker-compose-*.yml files.

```
docker-compose -f docker-compose-demo.yml build
docker-compose -f docker-compose-demo.yml up
```

### Architectural Considerations ##
##### Message Bus - one-to-many communication #####
A message bus for communication allows us to loosely couple services in a one-to-many manner. Neither at the time of writing a service, nor at any time in production does a service have to know, who will consume its data.

##### Lightweight Message Bus #####
We have chosen **MQTT publish and subscribe** message bus protocol, as it is lightweight and MQTT broker implementations are performant and scalable. Moreover it provides both a TCP socket and a Websocket interface.

##### Minimal MQTT Featureset ####
We are only utilizing QOS 0, as we follow an non deterministic approach to message delivery in our own service architecture. We count on messages being lost and account for this on a different architectural level, thereby making out infrastructure more tolerant to errors. Additionally, only using the most basic MQTT features makes us more protocol independant, as this minimal feature set is supported in a wide range of other message bus protocols as well.

##### Convention over Configuration #####
We follow a Convention over Configuration approach. Configuration files - such as subscription.txt - and also the processor file need to reside in certain locations to be started without further configuration needed. Configuration is only necessary, if you deviate from the norm.

### Development Goals ###
* Adding interfaces for other message bus protocols
