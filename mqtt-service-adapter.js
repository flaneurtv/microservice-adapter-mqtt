/**
 * require modules
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
var mqtt_listener_credentials_file = '/run/secrets/mqtt_listener.json';
var mqtt_publisher_credentials_file = '/run/secrets/mqtt_publisher.json';

/**
 * importing and set all necessary ENV vars
 */
var debug = process.env.DEBUG || false; // set DEBUG to true/false in ENV to log messages
var log_levels = ["emergency","alert","critical","error","warning","notice","info","debug"]
var log_level = process.env.LOG_LEVEL || "error"
if (debug === "true") { log_level = "debug" }
var bridge = process.env.BRIDGE || false; // set BRIDGE to true in ENV to bridge messages from MQTT_LISTENER_URL to MQTT_PUBLISHER_URL
var namespace = process.env.NAMESPACE || "default";
var namespace_listener = process.env.NAMESPACE_LISTENER || namespace;
var namespace_publisher = process.env.NAMESPACE_PUBLISHER || namespace;
process.env.NAMESPACE_LISTENER = namespace_listener
process.env.NAMESPACE_PUBLISHER = namespace_publisher
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
 * Creates MQTT_LISTENER connection with credentials if available
 */
var mqtt_listener_credentials = {username: "", password: ""}
if (fs.existsSync(mqtt_listener_credentials_file)) { mqtt_listener_credentials = require(mqtt_listener_credentials_file) }
// Connection for MQTT bus listener
var listener_client_id = service_name + "_" + service_host + "_" + service_uuid + "_listener"
var listener_will_object = generate_debug_message("INFO", "last will", listener_client_id)
var mqtt_listener = mqtt.connect(mqtt_listener_url_object, {
    username: mqtt_listener_credentials.username,
    password: mqtt_listener_credentials.password,
	clientId: listener_client_id,
    will: {
		topic: listener_will_object.topic,
		payload: listener_will_object.toString()
    }
});

/**
 * Creates MQTT_PUBLISHER connection with credentials if available
 * If LISTENER and PUBLISHER connection URL and credentials are equal,
 * it will just use the LISTENER connection.
 */
var mqtt_publisher_credentials = {username: "", password: ""}
if (fs.existsSync(mqtt_publisher_credentials)) { mqtt_publisher_credentials = require(mqtt_publisher_credentials) }
if (mqtt_listener_url_object.href === mqtt_publisher_url_object.href && mqtt_listener_credentials == mqtt_publisher_credentials) {
    var mqtt_publisher = mqtt_listener;
	logme("DEBUG", "MQTT connection", "listener and publisher are equal")
} else {
	var publisher_client_id = service_name + "_" + service_host + "_" + service_uuid + "_publisher"
	var publisher_will_object = generate_debug_message("INFO", "last will", publisher_client_id)
    var mqtt_publisher = mqtt.connect(mqtt_publisher_url_object, {
        username: mqtt_publisher_credentials.username,
        password: mqtt_publisher_credentials.password,
		clientId: publisher_client_id,
        will: {
            topic: publisher_will_object.topic,
            payload: publisher_will_object.toString()
        }
    });
}

/**
 * Starting processor,
 * creating pipes to it and setting event handlers for them
 * but only, if we are not in BRIDGE mode
 */
if (bridge !== "true") {
	logme("info", "Spawning processor", service_processor);
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
		line = line.trim();
		try {
			processor_stdout_message = JSON.parse(line);
			if (mqtt_publisher.connected === true) {
		        mqtt_publisher.publish(processor_stdout_message.topic, line);
		    }
			logme("debug", "processor_stdout_message", line);
		}
		catch(err) {
	        logme('error', err + ' in processor_stdout_message', line);
	    }
	});

	// FIXME: seems to never trigger at all, even if I kill -SIGINT the
	// processor and even though processor.on('exit') shows SIGINT as signal
	processor_stdout.on('SIGINT', () => { logme('INFO', "Processor STDOUT readline SIGINT event emitted"); });
	processor_stdout.on('close', () => { logme('INFO', "Processor STDOUT readline CLOSE event emitted"); });

	// Processor stderr is forwarded to out logme facility
	// QUESTION: why is the line !== null here?
	processor_stderr.on('line', (line) => {
	    if (line !== null) {
	        line = line.trim();
	        logme('error','processor_stderr_message', line);
	    }
	});
	processor_stderr.on('SIGINT', () => { logme('INFO', "Processor STDERR readline SIGINT event emitted"); });
	processor_stderr.on('close', () => { logme('INFO', "Processor STDERR readline CLOSE event emitted"); });

	/**
	 * Processor event handlers
	 * Events are triggered in this order
	 * processor.on('exit')
	 * processor.on('close')
	 */
	 processor.on('exit', (code, signal) => {
	     logme("info", "Processor EXIT event emitted", 'code ' + code + ' and signal ' + signal);
		 logme("info", "Processor.kill() initiated");
	     processor.kill();
	 });

	processor.on('close', (code, signal) => {
	    logme("info", "Processor CLOSE event emitted", 'code ' + code + ' and signal ' + signal);
		exit_gracefully(0)
	});

	processor.on("error", function(error) {
	    logme("info", "Processor ERROR event emitted", 'code ' + code + ' and signal ' + signal);
	    processor.kill();
	});

}

/**
 * MQTT Publisher event handlers
 */
mqtt_listener.on("offline", function() { logme('info','MQTT Listener offline', mqtt_listener_url_object.href) });
mqtt_listener.on("reconnect", function() { logme('info', 'MQTT Listener trying to reconnect to', mqtt_listener_url_object.href) });

mqtt_listener.on("message", function(topic_in, message_in) {
	var topic, message
    logme('debug','MQTT_MESSAGE_RECEIVED', {topic: topic_in , message: message_in.toString()};
	/* if we run the the service-adapter as a message bridge between two MQTT
	buses or different namespaces, we do not pipe these messages through an
	external processor but instead rewrite the topic within the service-adapter
	code for performance reasons.
	*/
	if (bridge === "true") {
		// Alters message and topic and sends it directly to publisher
		if (namespace_listener !== namespace_publisher) {
			topic = topic_in.replace(namespace_listener, namespace_publisher)
			// FIXME: JSON.parse is probably expensive, better operate on message_string not message_object
			var message_object = JSON.parse(message_in.toString())
			message_object.topic = topic
			message = JSON.stringify(message_object)
		}
		if (mqtt_publisher.connected === true) {
	        mqtt_publisher.publish(topic, message);
			logme("debug","MQTT message relayed through bridge", topic_in + " => " + topic)
	    } else { logme("debug","MQTT message for bridge dropped", message) }
	} else {
		// Forwards message to processor
		try {
			message = JSON.stringify(JSON.parse(message_in.toString()))
			processor.stdin.write(message + '\n');
		} catch (e) {
			logme("error", e, message_in.toString())
		}
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
            if (line !== null && namespace_listener !== "null") {
                // Subscribe to topic it uses namespace variable, subscriptions.txt file SHOULD NOT have the namespace defined.
                mqtt_listener.subscribe(namespace_listener + '/' + line);
                logme("info","Topic subscribed", namespace_listener + '/' + line);
            }
        });
    } else {
        logme("info","subscriptions.txt file is missing, NO TOPICS ARE SUBSCRIBED.");
    }
});

// Prints error messages on the MQTT listener bus
mqtt_listener.on("error", function(error) {
    logme("error","MQTT Listener ERROR", error);
    if (error.code === 5) { // Error: MQTT Connection refused: Not authorized
        exit_gracefully(0);
    }
});

/**
 * MQTT Publisher event handlers
 */
mqtt_publisher.on("connect", (connack) => { logme('info','MQTT publisher connected to', mqtt_publisher_url_object.href) });
mqtt_publisher.on("offline", function() { logme('info', 'MQTT Publisher offline', mqtt_publisher_url_object.href) });
mqtt_publisher.on("reconnect", function() { logme("info", "MQTT Publisher trying to reconnect to", mqtt_publisher_url_object.href) });
mqtt_publisher.on("error", function(error) {
    logme("error","ERROR publisher", error);
    if (error.code === 5) { // Error: MQTT Connection refused: Not authorized
        exit_gracefully(0)
    }
});

// Handles uncaughtException errors node exits gracefully on processor termination
process.on('uncaughtException', function(error) {
	logme("error", "uncaughtException", error);
	exit_gracefully(0);
});

function exit_gracefully(exit_code=0){
	setTimeout(function () {
		mqtt_publisher.end();
		mqtt_listener.end();
		process.exit(exit_code);
	}, 100);
}

// Prints in console and publish on the MQTT bus log.
// FIXME: ERROR messages need to go to stderr not stdout
function logme (level, description, message="") {
	log_level_index = log_levels.findIndex(function(element, index, array) { return element === log_level });
	message_level_index = log_levels.findIndex(function(element, index, array) { return element === level });
    if (message_level_index <= log_level_index) {
		log_object = generate_debug_message(level, description, message)
		if (message_level_index <= 3) { process.stderr.write(level + ' => ' + description + ': ' + message + '\n') }
		else { console.log(level + ' => ' + description + ': ' + message); }
        if (mqtt_publisher.connected === true) {
            mqtt_publisher.publish(log_object.topic, log_object.toString());
        }
    }
}

function generate_debug_message(level, description, message){
	var message_object = {
		topic: namespace_publisher + "/" + "log/" + level,
		service_name: service_name,
		service_uuid: service_uuid,
		service_host: service_host,
		payload: {
			log_message: {
				log_level: level,
				description: description,
				body: message
			}
		}
	}
	return message_object
}
