// Setup variables
var url = require('url');
var uuid = require('uuid');
var mqtt = require('mqtt');
var readline = require('readline');
var fs = require('fs');
var child_process = require('child_process');

// importing all necessary ENV vars
var namespace = process.env.NAMESPACE || ""
var service_processor = process.env.SERVICE_PROCESSOR || "./service-processor/processor.sh"
// TODO: can we read the SERVICE_NAME from package.json?
// Apparently a json file can be read with 'require'
var service_name = process.env.SERVICE_NAME || "null"
var service_uuid = process.env.SERVICE_UUID || uuid() // randomly assigned
var service_host = process.env.SERVICE_HOST || os.hostname() // equals the docker container ID
var mqtt_listener_url_object = url.parse(process.env.MQTT_LISTENER_URL || "tcp://mqtt:1883");
var mqtt_publisher_url_object = url.parse(process.env.MQTT_PUBLISHER_URL || "tcp://mqtt:1883");
// TODO: I liked the idea of reading this from a textfile. We could add a ./service-processor/subscriptions.txt
var mqtt_subscriptions = process.env.MQTT_SUBSCRIPTIONS || "";

// Prints MQTT listener/publisher url object
// console.log('info: mqtt_listener_url parsed:' + JSON.stringify(mqtt_listener_url_object));
// console.log('info: mqtt_publisher_url parsed:' + JSON.stringify(mqtt_publisher_url_object));

// Connection for MQTT bus listener
var mqtt_listener = mqtt.connect(mqtt_listener_url_object, {
    // username: "type1tv",
    // password: "nuesse",
    // will: {
    //     topic: namespace + "log",
    //     payload: "{service: " + service_name + ", event: 'last will'}"
    // }
});

// Connection for MQTT bus publisher
if (mqtt_listener_url_object.href === mqtt_publisher_url_object.href) {
    var mqtt_publisher = mqtt_listener;
} else {
    var mqtt_publisher = mqtt.connect(mqtt_publisher_url_object, {
        // username: "type1tv",
        // password: "nuesse",
        // will: {
        //     topic: namespace + "log",
        //     payload: "{service: " + service_name + ", event: 'last will'}"
        // }
    });
}


/**
 * Events for MQTT listener
 */

// Listen to messages on the MQTT bus
mqtt_listener.on("message", function(topic, message) {
    console.log('event => MQTT_MESSAGE_RECEIVED, topic: "' + topic + '", message: "' + message.toString().trim() + '"');

    // Forwards message to processor
    // FIXME: the child prosess is not spawned on every message but is a long
    // living process with an internal loop reading from stdin line by line.
    // When the process dies, the service-adapter needs to die with it!
    sp = child_process.spawn(service_processor);
    sp.stdin.write(message.toString().trim() + '\n');

    // Processor stdout response, publish response
    // FIXME: I am pretty sure on_data does not suffice here. Because data does
    // NOT mean line. It is just the current buffer. So it can be half a line or
    // it can be two lines!!! You have to separate things at the newline symbol.
    // There is no action until a full line has been received ergo until a
    // newline has been received. This is why I used readline here!!!
    sp.stdout.on('data', function (data) {
        // console.log(data.toString());
        processor_stdout_message = JSON.parse(data);
        mqtt_publisher.publish(processor_stdout_message.topic, data.toString());
    });

});

// Prints when connected to MQTT listener then makes subscriptions
mqtt_listener.on("connect", (connack) => {
    console.log('event => MQTT listener connected to: "' + mqtt_listener_url_object.href + '"');
    // Checks subscriptions of ENV
    if (mqtt_subscriptions !== null) {
        // Splits subscriptions by ; and then subscribes
        var arr = mqtt_subscriptions.split(";");
        for(var i = 0; i < arr.length; i++) {
            mqtt_listener.subscribe(arr[i]);
            console.log('event => Topic subscribed: "' + arr[i] + '"');
        }
    }
});

// Prints error messages on the MQTT listener bus
mqtt_listener.on("error", function(error) {
    console.log("event => ERROR listener: ", error);
});

// Prints when MQTT listener bus is offline
mqtt_listener.on("offline", function() {
    console.log('event => MQTT Listener Server offline: "' + mqtt_listener_url_object.href + '"');
});

// Prints when MQTT listener had to reconnect
mqtt_listener.on("reconnect", function() {
    console.log('event => Trying to reconnect to listener in: "' + mqtt_listener_url_object.href + '"');
});


/**
 * Events for MQTT publisher
 */

// Prints when connected to MQTT publisher
mqtt_publisher.on("connect", (connack) => {
    console.log('event => MQTT publisher connected to: "' + mqtt_publisher_url_object.href + '"');
});

// Prints error messages on the MQTT bus
mqtt_publisher.on("error", function(error) {
    console.log("event => ERROR publisher: ", error);
});

// Prints when MQTT bus is offline
mqtt_publisher.on("offline", function() {
    console.log('event => MQTT Publisher Server offline: "' + mqtt_publisher_url_object.href + '"');
});

// Prints when MQTT had to reconnect
mqtt_publisher.on("reconnect", function() {
    console.log('event => Trying to reconnect to publisher in: "' + mqtt_publisher_url_object.href + '"');
});

// FIXME: this is supposed to be part of the processor not the service-adapter
// Sends tick test every 3 seconds (3000ms)
function sendTick() {
  // FIXME: the "flaneur" part of the topic has to come from the NAMESPACE env
    mqtt_publisher.publish("flaneur/tick", JSON.stringify(
        {
            "topic": "flaneur/tick",
            "message": "tick"
        })
    );
}
setInterval(sendTick, 3000);

// FIXME: you could remodle this into the "alive" message functionality I was
// talking about. This "alive" responder, which responds to the "tick" messages
// does really belong to the service-adapter code. The tick_responder test case
// service-processor does not.
// Sends test after 1 second (1000ms)
setTimeout(function() {
    mqtt_publisher.publish("flaneur/tusd/upload_success", JSON.stringify(
        {
            "topic": "flaneur/tusd/upload_success",
            "service_uuid": "SERVICEUUID",
            "service_name": "SERVICENAME",
            "service_host": "SERVICEHOST",
            "created_at": "CREATEDAT",
            "payload": {
                "tick_uuid": "TICKUUID"
            }
        })
    );
}, 1000);
