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
var mqtt_listener_credentials = '/run/secrets/mqtt_listener.json';
var mqtt_publisher_credentials = '/run/secrets/mqtt_publisher.json';

/**
 * importing all necessary ENV vars
 */
var debug = process.env.DEBUG || true; // set DEBUG to true/false in ENV to log messages
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
 * checks existing credentials for MQTT listener
 */
if (fs.existsSync(mqtt_listener_credentials)) {
    // Parse JSON
    mqtt_listener_credentials = require(mqtt_listener_credentials);
    // Connection for MQTT bus listener
    var mqtt_listener = mqtt.connect(mqtt_listener_url_object, {
        username: mqtt_listener_credentials.username,
        password: mqtt_listener_credentials.password,
        will: {
            topic: namespace + '/' + "log",
            payload: "{service: " + service_name + ", event: 'last will'}"
        }
    });
} else {
    // Connection for MQTT bus listener WITHOUT AUTHORIZATION
    var mqtt_listener = mqtt.connect(mqtt_listener_url_object, {
        will: {
            topic: namespace + '/' + "log",
            payload: "{service: " + service_name + ", event: 'last will'}"
        }
    });
}

/**
 * checks existing credentials for MQTT publisher
 */
if (fs.existsSync(mqtt_publisher_credentials)) {
    mqtt_publisher_credentials = require(mqtt_publisher_credentials); // Parse JSON
    // Connection for MQTT bus publisher
    if (mqtt_listener_url_object.href === mqtt_publisher_url_object.href) {
        var mqtt_publisher = mqtt_listener;
    } else {
        var mqtt_publisher = mqtt.connect(mqtt_publisher_url_object, {
            username: mqtt_publisher_credentials.username,
            password: mqtt_publisher_credentials.password,
            will: {
                topic: namespace + '/' + "log",
                payload: "{service: " + service_name + ", event: 'last will'}"
            }
        });
    }
} else {
    // Connection for MQTT bus listener WITHOUT AUTHORIZATION
    var mqtt_publisher = mqtt.connect(mqtt_publisher_url_object, {
        will: {
            topic: namespace + '/' + "log",
            payload: "{service: " + service_name + ", event: 'last will'}"
        }
    });
}

/**
 * starts the processor
 */
console.log('info => spawning processor: ' + service_processor);
var processor;
// if processor is a javascript file, then use exec, otherwise use spawn
if (service_processor.substr(service_processor.lastIndexOf('.') + 1) === 'js') {
    processor = child_process.exec(service_processor, { env: process.env});
} else {
    processor = child_process.spawn(service_processor, { env: process.env});
}
const processor_stdout = readline.createInterface({ input: processor.stdout});
const processor_stderr = readline.createInterface({ input: processor.stderr});

/**
 * Processor stdout is published on MQTT if connected and valid JSON
 * If MQTT is not connected, lines are dropped to avoid late messages
 */
processor_stdout.on('line', (line) => {
	json_valid = false;
	try {
		line = line.trim();
		processor_stdout_message = JSON.parse(line);
		json_valid = true;
	}
	catch(err) {
        console.log('error => ' + err + ', processor_stdout_message => ' + line);
    }
    if (json_valid === true && mqtt_publisher.connected === true) {
        mqtt_publisher.publish(processor_stdout_message.topic, line);
        logme('{"service": "' + service_name + '", "event": "processor_stdout_message => ", "message": "' + line + '"}');
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
 * On CLOSE processor_stdout
 */
// FIXME: define proper log messages to publish on MQTT
processor_stdout.on('close', () => {
    logme('{"service": "' + service_name + '", "event": "Readline CLOSE event emitted", "reaction": "termination"}');
    console.log('event => Readline CLOSE event emitted');
});

/**
 * On SIGINT processor
 */
// FIXME: define proper log messages to publish on MQTT
// seems to never trigger at all, even if I kill -SIGINT the processor and
// even though processor.on('exit') shows SIGINT as signal
processor_stdout.on('SIGINT', () => {
    logme('{"service": "' + service_name + '", "event": "Readline SIGINT event emitted", "reaction": "termination"}');
    console.log('event => Readline SIGINT event emitted');
});

/**
 * Handles close on processor
 */
processor.on('close', (code, signal) => {
    // triggers after processor.on('exit')
    logme('{"service": "' + service_name + '", "event": "Processor CLOSE event emitted", "reaction": "termination"}');
    console.log('event => Processor CLOSE with code ' + code + ' and signal ' + signal);
    setTimeout(function () {
        mqtt_publisher.end();
        mqtt_listener.end();
        process.exit(0);
    }, 100);
});

/**
 * Handles exit on processor
 */
processor.on('exit', (code, signal) => {
    // triggers before processor.on('close')
    logme('{"service": "' + service_name + '", "event": "Processor EXIT event emitted", "reaction": "termination"}');
    console.log('event => Processor EXIT with code ' + code + ' and signal ' + signal);
    processor.kill();
});

/**
 * Handles errors on processor
 */
processor.on("error", function(error) {
    // triggers before processor.on('error')
    logme('{"service": "' + service_name + '", "event": "Processor ERROR event emitted", "reaction": "termination"}');
    console.log('event => Processor ERROR : ' + error);
    processor.kill();
});

/**
 * Handles uncaughtException errors node exits gracefully on processor termination
 */
process.on('uncaughtException', function(error) {
    console.log('event => uncaughtException => ' + error);
    process.exit(0);
});

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
    if (error.code === 5) { // Error: MQTT Connection refused: Not authorized
        process.exit(0);
    }
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
    if (error.code === 5) { // Error: MQTT Connection refused: Not authorized
        process.exit(0);
    }
});

// Prints when MQTT bus is offline
mqtt_publisher.on("offline", function() {
    console.log('event => MQTT Publisher Server offline: "' + mqtt_publisher_url_object.href + '"');
});

// Prints when MQTT had to reconnect
mqtt_publisher.on("reconnect", function() {
    console.log('event => Trying to reconnect to publisher in: "' + mqtt_publisher_url_object.href + '"');
});

// Prints in console and publish on the MQTT bus log.
function logme (message) {
    if (debug === true) {
        console.log('event => Logme triggered: ' + message);
        if (mqtt_publisher.connected === true) {
            mqtt_publisher.publish(namespace + '/' + 'log', message);
        }
    }
}
