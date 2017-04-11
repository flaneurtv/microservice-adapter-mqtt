// Setup variables
var url = require('url');
var uuid = require('uuid/v1');
var mqtt = require('mqtt');
var readline = require('readline');
var fs = require('fs');
var os = require('os');
var moment = require('moment');
var child_process = require('child_process');
var pjson = require('../package.json');
var subscriptions_txt = './service-processor/subscriptions.txt';

// importing all necessary ENV vars
var namespace = process.env.NAMESPACE || "";
var service_name = pjson.name; // Name of service comes from package.json
var service_uuid = uuid(); // randomly assigned
var service_host = os.hostname(); // equals the docker container ID
var service_processor = "./service-processor/processor";
var mqtt_listener_url_object = url.parse(process.env.MQTT_LISTENER_URL || "tcp://mqtt:1883");
var mqtt_publisher_url_object = url.parse(process.env.MQTT_PUBLISHER_URL || "tcp://mqtt:1883");

// start the processor
console.log('info: spawning processor: ' + service_processor);
const processor = child_process.spawn(service_processor, []);
const processor_stdout = readline.createInterface({ input: processor.stdout});
const processor_stderr = readline.createInterface({ input: processor.stderr});

// Processor stdout is published on MQTT if connected and valid JSON
// If MQTT is not connected, lines are dropped to avoid late messages
processor_stdout.on('line', (line) => {
    if (JSON.parse(line) && mqtt_publisher.connected === true) {
        line = line.trim();
        processor_stdout_message = JSON.parse(line);
        mqtt_publisher.publish(processor_stdout_message.topic, line);
        console.log('processor_stdout_message: ' + line);
    }
});

// Processor stderr is forwarded to log line by line
// QUESTION: I am unsure whether to forward to parent stderr or to stdout via
// log facility
processor_stderr.on('line', (line) => {
    if (line !== null) {
        line = line.trim();
        console.log('processor_stderr_message: ' + line);
    }
});

// FIXME: define proper log messages to publish on MQTT
// FIXME: make sure the adapter terminates when the processor is killed
processor_stdout.on('close', () => {
    console.log('event: Readline CLOSE event emitted');
    if (mqtt_publisher.connected === true) {
        mqtt_publisher.publish(namespace + 'log', '{"service": "status_checker", "event": "readline CLOSE event emitted", "reaction": "terminating"}');
    }
    processor_stdout.close();
    mqtt_publisher.end();
    mqtt_listener.end();
});

// FIXME: define proper log messages to publish on MQTT
// FIXME: make sure node exits gracefully on processor termination
processor_stdout.on('SIGINT', () => {
    console.log('event: Readline SIGINT event emitted');
    if (mqtt_publisher.connected === true) {
        mqtt_publisher.publish(namespace + 'log', '{"service": "status_checker", "event": "readline SIGINT event emitted", "reaction": "terminating"}');
    }
    processor_stdout.close();
    mqtt_publisher.end();
    mqtt_listener.end();
});

processor.on('close', (code) => {
    console.log(`child process exited with code ${code}`);
});

// Prints MQTT listener/publisher url object
// console.log('info: mqtt_listener_url parsed:' + JSON.stringify(mqtt_listener_url_object));
// console.log('info: mqtt_publisher_url parsed:' + JSON.stringify(mqtt_publisher_url_object));

// Connection for MQTT bus listener
// FIXME: the login credentials shall be read from
// /run/secrets/mqtt_publisher.json and
// /run/secrets/mqtt_listener.json
// If you have a more meaningful name, please tell me.
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
    if (processor.connected) {
        processor.stdin.write(message.toString().trim() + '\n');
    }

});

// Prints when connected to MQTT listener then makes subscriptions
mqtt_listener.on("connect", (connack) => {
    console.log('event => MQTT listener connected to: "' + mqtt_listener_url_object.href + '"');

    // checks if subscriptions.txt files exists
    if (fs.existsSync(subscriptions_txt)) {
        // Reads file subscriptions.txt line by line for MQTT topics the adapter should be subscribed to
        var rl = readline.createInterface({
            input: fs.createReadStream(subscriptions_txt)
        });
        // With each line of subscriptions.txt create a subscription for the adapter
        rl.on('line', function(line) {
            // checks it´s not an empty line
            if (line !== null && namespace !== "null") {
                // Subscribe to topic it uses namespace variable, subscriptions.txt file SHOULD NOT have the namespace defined.
                mqtt_listener.subscribe(namespace + line);
                console.log("event => Topic subscribed: " + namespace + line);
            }
        });
    } else {
        console.log("event => subscriptions.txt file is missing, NO TOPICS ARE SUBSCRIBED.");
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
// function sendTick() {
//   // FIXME: the "flaneur" part of the topic has to come from the NAMESPACE env
//     mqtt_publisher.publish(namespace + "/tick", JSON.stringify(
//         {
//             "topic": namespace + "/tick",
//             "message": "tick",
//             "created_at": moment().utc().unix(),
//         })
//     );
// }
// setInterval(sendTick, 3000);

// FIXME: you could remodle this into the "alive" message functionality I was
// talking about. This "alive" responder, which responds to the "tick" messages
// does really belong to the service-adapter code. The tick_responder test case
// service-processor does not.
// Sends test after 1 second (1000ms)
// setTimeout(function() {
//     mqtt_publisher.publish(namespace + "/tusd/upload_success", JSON.stringify(
//         {
//             "topic": namespace + "/tusd/upload_success",
//             "service_uuid": service_uuid,
//             "service_name": service_name,
//             "service_host": service_host,
//             "created_at": moment().utc().unix(),
//             "payload": {
//                 "tick_uuid": "TICKUUID"
//             }
//         })
//     );
// }, 1000);
