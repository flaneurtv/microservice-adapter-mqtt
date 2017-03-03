var url = require('url');
var uuid = require('uuid');
var mqtt = require('mqtt');
var fs = require('fs');

// importing all necessary ENV vars
var namespace = process.env.NAMESPACE || ""
var service_processor = process.env.SERVICE_PROCESSOR || "./processor.sh"
var service_name = process.env.SERVICE_NAME || "null"
var service_uuid = process.env.SERVICE_UUID || ""
var service_host = process.env.SERVICE_HOST || ""
var mqtt_listener_url_object = url.parse(process.env.MQTT_LISTENER_URL || "tcp://mqtt:1883")
var mqtt_publisher_url_object = url.parse(process.env.MQTT_PUBLISHER_URL || "tcp://mqtt:1883")

console.log('info: mqtt_listener_url parsed:' + JSON.stringify(mqtt_listener_url_object));
console.log('info: mqtt_publisher_url parsed:' + JSON.stringify(mqtt_publisher_url_object));

const mqtt_listener = mqtt.connect(mqtt_listener_url_object, {username: "type1tv", password: "nuesse", will: {topic: namespace+"log", payload: "{service: "+service_name+", event: 'last will'}"}});
const mqtt_publisher = mqtt.connect(mqtt_publisher_url_object, {username: "type1tv", password: "nuesse", will: {topic: namespace+"log", payload: "{service: "+service_name+", event: 'last will'}"}});

//console.log('info: mqtt_listener:' + JSON.stringify(mqtt_listener));
//console.log('info: mqtt_publisher:' + JSON.stringify(mqtt_publisher));

mqtt_listener.on('connect', (connack) => {
  console.log('event: MQTT_LISTENER connected');
});

mqtt_publisher.on('connect', (connack) => {
  console.log('event: MQTT_PUBLISHER connected');
});

console.log('spawning processor: '+service_processor);
const child_process = require('child_process');
const processor = child_process.spawn(service_processor, []);

const readline = require('readline');
const rl = readline.createInterface({ input: processor.stdout});

rl.on('line', (line) => {
  if (line !== null && mqtt_publisher.connected == true) {
    // FIXME: line needs to be split into topic and message and then published accordingly
    protocol_message=JSON.parse(line);
    mqtt_publisher.publish(protocol_message["topic"], JSON.stringify(protocol_message["message"]));
    console.log('protocol_message: '+line.trim());
  }
});

rl.on('close', () => {
  console.log('event: Readline CLOSE event emitted');
  if (mqtt_publisher.connected == true) {
    mqtt_publisher.publish(namespace+'log', '{"service": "status_checker", "event": "readline CLOSE event emitted", "reaction": "terminating"}')
  }
  rl.close();
  mqtt_publisher.end();
});

rl.on('SIGINT', () => {
  console.log('event: Readline SIGINT event emitted');
  if (mqtt_publisher.connected == true) {
    mqtt_publisher.publish(namespace+'log', '{"service": "status_checker", "event": "readline SIGINT event emitted", "reaction": "terminating"}')
  }
  rl.close();
  mqtt_publisher.end();
});

processor.stderr.on('data', (data) => {
  console.log(`stderr: ${data}`);
});

processor.on('close', (code) => {
  console.log(`child process exited with code ${code}`);
});

// FIXME: we need to subscribe to all topics specified via ENV and the tick and forward it to the processor
// we always subscribe to tick, but only forward tick messages, if it is part of the ENV subscription list
// for now I subscribe to a fixed topic for testing

mqtt_listener.subscribe(namespace+'test')
mqtt_listener.on('message', function (topic, message) {
  console.log('event: MQTT_MESSAGE_RECEIVED, topic: "' +topic+ '", message: '+ message.toString().trim());
  processor.stdin.write('{"topic":"'+topic+'","message":'+ message.toString().trim() +'}\n');
});
