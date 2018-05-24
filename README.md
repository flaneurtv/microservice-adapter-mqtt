# Service Adapter MQTT #

## What is a microservice ##
A microservice is a lean piece of software, which by concept is meant to be single-purpose - doing one job only and doing it well - which equates to the UNIX principle. 

For demonstration purposes, let's say we have a service, which tells the time when we ask for it. Let's call it time-service.

## MQTT message bus ##
This brings us to the second essential part of a microservice architecture, a means of communication with and amongst these services - the communication protocol. For our architecture we use a message bus for this purpose, specifically the **MQTT publish and subscribe** message bus protocol. A message consists of a topic, which other clients can subscribe to, and a body. The topic might have a tree structure and one can subscribe to any portion of that tree using wildcards.

This enables us to loosely couple services in a one-to-many manner, without being required to know, who exactly will consume the data we publish - neither at the time of writing the code, nor later.

In terms of our example above, we would publish a request on the message bus with the topic "what's the time" and the time-service, which subscribed to this topic, will respond with the topic "current time", carrying the actual time data in the message body.

Now here comes the magic, the time-service does directly address any recipients of this message - it does care who receives that data.

We might have a display-time-service subscribed to the "current time" topic, updating our website on every "current time" message.

But we might also have a big-surprise-service subscribed to the "current time", reading this value and setting off the bomb, when current time == deadline.

## Service Adapter MQTT ##

Since all these services need to implement the same communication routines, it makes sense to split the communication part from the routines, which do the actual job. We call the latter the service-processor.

What comes to mind first is to put all communication functionality into a library or module. As a result however, we would restrict microservice developers to use the language we have chosen for the library. Or, if we decided to provide implementations in multiple languages, keeping things up-to-date would become increasingly difficult.

Therefore we have chosen to separate communication from functionality on a different level - the process level. A microservice using this software therefore consists of two separate runtimes, tied together via UNIX pipes.

* Service Processor (functionality)
* Service Adapter MQTT (communication)

### Service Processor ###

The service processor carries the routines, which are doing the actual work. In our example, that would be the code which determines the current time or triggers the bomb. This service processor has one input and two outputs like each UNIX process:
* stdin
* stdout
* stderr

Each line on input/output represents one message. These messages are formatted in JSON following the JSON-pipe-protocol, we defined for this purpose. This protocol is a generalization of a publish/subscribe message bus protocol.

### Service Adapter ###

The service adapter acts as a translator between the JSON-pipe-protocol and the actual message bus protocol used. Therefore you could switch out the service-adapter-mqtt with a service-adapter-socket.io and would not have to change a thing in the service-processor.

The service adapter is parametrized at start-up with a list of topics, it should subscribe to. It also implements message verification and error handling (and jwt securtiy features in the future).

#### To sum that up, here is what the service adapter does: ####
* starts up the service-processor as a long-running subprocess
* establishes pipes to the subprocesses stdin, stdout, stderr
* subscribes to MQTT topics
* receives MQTT messages
* forwards these messages to service-processors stdin
* receives responses on service-processors stdout
* publishes these responses to MQTT
* receives errors on service-processors stderr
* publishes these errors to predefined MQTT topic

#### Functional Diagram ####
```
  |  +----------------------------------------------------+
 M|  |  service-adapter-mqtt            subprocess        |
 Q|  |                           +---------------------+  |
 T|  |                           |  service-processor  |  |
 T|  |                           |                     |  |
  +-----> subscribe           ------> stdin            |  |
  +-----< publish             <------ stdout           |  |
 B|  |                        <------ stderr           |  |
 U|  |                           +---------------------+  |
 S|  +----------------------------------------------------+
  |
```

## TODO ##

For rapid prototyping service-adapter-mqtt has been implemented in node.js. This has a major disadvantage for deployment, as we can not generate a binary and therefore have to provide a full runtime environment for the adapter. When the service-processor is implemented in another language, crafting docker containers for the microservice is becoming quite cumbersome. Therefore we want to re-implement this prototype in Golang. The service-adapter then would be only one single binary, which can easily be added to any runtime container with a simple `COPY from=service-adapter` Dockerfile instruction utilizing multistage builds.