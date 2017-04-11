// Setup variables
var url = require('url');
var uuid = require('uuid');
var mqtt = require('mqtt');
var readline = require('readline');
var fs = require('fs');
var child_process = require('child_process');
var readline = require('readline');

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

var subscriptions_txt = './deploy/subscriptions.txt';

// start the processor
console.log('info: spawning processor: ' + service_processor);
const processor = child_process.spawn(service_processor, []);
const processor_stdout = readline.createInterface({ input: processor.stdout});
const processor_stderr = readline.createInterface({ input: processor.stderr});

// Processor stdout is published on MQTT if connected and valid JSON
// If MQTT is not connected, lines are dropped to avoid late messages
processor_stdout.on('line', (line) => {
  if (JSON.parse(line) && mqtt_publisher.connected == true) {
    line = line.trim();
    processor_stdout_message = JSON.parse(line);
    mqtt_publisher.publish(processor_stdout_message["topic"], line);
    console.log('processor_stdout_message: ' + line;
  }
});

// Processor stderr is forwarded to log line by line
// QUESTION: I am unsure whether to forward to parent stderr or to stdout via
// log facility
processor_stderr.on('line', (line) => {
  if (line !== null) {
    line = line.trim();
    console.log('processor_stderr_message: ' + line;
  }
});

// FIXME: define proper log messages to publish on MQTT
rl.on('close', () => {
  console.log('event: Readline CLOSE event emitted');
  if (mqtt_publisher.connected == true) {
    mqtt_publisher.publish(namespace+'log', '{"service": "status_checker", "event": "readline CLOSE event emitted", "reaction": "terminating"}')
  }
  rl.close();
  mqtt_publisher.end();
  mqtt_listener.end();
});

// FIXME: define proper log messages to publish on MQTT
// FIXME: make sure node exits gracefully on processor termination
rl.on('SIGINT', () => {
  console.log('event: Readline SIGINT event emitted');
  if (mqtt_publisher.connected == true) {
    mqtt_publisher.publish(namespace+'log', '{"service": "status_checker", "event": "readline SIGINT event emitted", "reaction": "terminating"}')
  }
  rl.close();
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

// Reads file subscriptions.txt line by line for MQTT topics the adapter should be subscribed to
var rl = readline.createInterface({
    input: fs.createReadStream(subscriptions_txt)
});

// With each line of subscriptions.txt create a subscription for the adapter
rl.on('line', function(line) {
    // checks it´s not an empty line
    if (line !== '') {
        // Subscribe to topic
        mqtt_listener.subscribe(line);
        console.log("event => Topic subscribed: " + line);
    }
});
