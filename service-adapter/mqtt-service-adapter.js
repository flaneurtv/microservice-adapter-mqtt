/**
 * Setup variables
 */
var url = require('url');
var uuid = require('uuid/v1');
var mqtt = require('mqtt');
var readline = require('readline');
var fs = require('fs');
var os = require('os');
var child_process = require('child_process');
var pjson = require('../package.json');
var subscriptions_txt = './service-processor/subscriptions.txt';

/**
 * importing all necessary ENV vars
 */
var namespace = process.env.NAMESPACE || "default";
var service_name = process.env.SERVICE_NAME || pjson.name; // Name of service comes from package.json
var service_uuid = uuid(); // randomly assigned
var service_host = os.hostname(); // equals the docker container ID
process.env.SERVICE_NAME = service_name;
process.env.SERVICE_UUID = service_uuid;
process.env.SERVICE_HOST = service_host;
var service_processor = process.env.SERVICE_PROCESSOR || "./service-processor/processor";
var mqtt_listener_url_object = url.parse(process.env.MQTT_LISTENER_URL || "tcp://mqtt:1883");
var mqtt_publisher_url_object = url.parse(process.env.MQTT_PUBLISHER_URL || "tcp://mqtt:1883");

/**
 * starts the processor
 */
console.log('info => spawning processor: ' + service_processor);
processor_env = Object.create( process.env );

const processor = child_process.spawn(service_processor, { env: process.env});
const processor_stdout = readline.createInterface({ input: processor.stdout});
const processor_stderr = readline.createInterface({ input: processor.stderr});

/**
 * Processor stdout is published on MQTT if connected and valid JSON
 * If MQTT is not connected, lines are dropped to avoid late messages
 */
processor_stdout.on('line', (line) => {
    if (JSON.parse(line) && mqtt_publisher.connected === true) {
        line = line.trim();
        processor_stdout_message = JSON.parse(line);
        mqtt_publisher.publish(processor_stdout_message.topic, line);
        console.log('processor_stdout_message => ' + line);
    }
});

/**
 * Processor stderr is forwarded to log line by line
 */
// QUESTION: I am unsure whether to forward to parent stderr or to stdout via log facility
processor_stderr.on('line', (line) => {
    if (line !== null) {
        line = line.trim();
        console.log('processor_stderr_message => ' + line);
    }
});

/**
 * On CLOSE processor
 */
// FIXME: define proper log messages to publish on MQTT
// FIXME: make sure the adapter terminates when the processor is killed
processor_stdout.on('close', () => {
    console.log('event: Readline CLOSE event emitted');
    if (mqtt_publisher.connected === true) {
        mqtt_publisher.publish(namespace + '/' + 'log', '{"service": "status_checker", "event": "readline CLOSE event emitted", "reaction": "terminating"}');
    }
    processor_stdout.close();
    mqtt_publisher.end();
    mqtt_listener.end();
});

/**
 * On SIGINT processor
 */
// FIXME: define proper log messages to publish on MQTT
// FIXME: make sure node exits gracefully on processor termination
processor_stdout.on('SIGINT', () => {
    console.log('event: Readline SIGINT event emitted');
    if (mqtt_publisher.connected === true) {
        mqtt_publisher.publish(namespace + '/' + 'log', '{"service": "status_checker", "event": "readline SIGINT event emitted", "reaction": "terminating"}');
    }
    processor_stdout.close();
    mqtt_publisher.end();
    mqtt_listener.end();
});

processor.on('close', (code) => {
    console.log('child process exited with code: ' + code);
});

/**
 * Prints MQTT listener/publisher url object
 */
// console.log('info: mqtt_listener_url parsed:' + JSON.stringify(mqtt_listener_url_object));
// console.log('info: mqtt_publisher_url parsed:' + JSON.stringify(mqtt_publisher_url_object));

/**
 * Connection for MQTT bus listener
 */
// FIXME: the login credentials shall be read from
// /run/secrets/mqtt_publisher.json and
// /run/secrets/mqtt_listener.json
// If you have a more meaningful name, please tell me.
var mqtt_listener = mqtt.connect(mqtt_listener_url_object, {
    // username: "type1tv",
    // password: "nuesse",
    // will: {
    //     topic: namespace + '/' + "log",
    //     payload: "{service: " + service_name + ", event: 'last will'}"
    // }
});

/**
 * Connection for MQTT bus publisher
 */
if (mqtt_listener_url_object.href === mqtt_publisher_url_object.href) {
    var mqtt_publisher = mqtt_listener;
} else {
    var mqtt_publisher = mqtt.connect(mqtt_publisher_url_object, {
        // username: "type1tv",
        // password: "nuesse",
        // will: {
        //     topic: namespace + '/' + "log",
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
    processor.stdin.write(message.toString().trim() + '\n');
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
                mqtt_listener.subscribe(namespace + '/' + line);
                console.log("event => Topic subscribed: " + namespace + '/' + line);
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
