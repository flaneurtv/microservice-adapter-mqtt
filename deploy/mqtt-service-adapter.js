// Setup variables
var url = require('url');
var uuid = require('uuid');
var mqtt = require('mqtt');
var readline = require('readline');
var fs = require('fs');
var mqtt_url_object = url.parse("tcp://mqtt:1883");
var subscriptions_txt = './deploy/subscriptions.txt';

// Prints MQTT url object
console.log('\n' + 'info => mqtt_url parsed:' + JSON.stringify(mqtt_url_object) + '\n');

// Connection for MQTT bus
const mqtt_client = mqtt.connect(mqtt_url_object, {
    // username: "type1tv",
    // password: "nuesse",
    // will: {
    //     topic: namespace + "log",
    //     payload: "{service: " + service_name + ", event: 'last will'}"
    // }
});

// Prints when connected to MQTT
mqtt_client.on('connect', (connack) => {
    console.log('event => MQTT connected');
});

// Prints error messages on the MQTT bus
mqtt_client.on("error", function(error) {
    console.log("event => ERROR: ", error);
});

// Prints when MQTT bus is offline
mqtt_client.on('offline', function() {
    console.log("event => MQTT Server offline");
});

// Prints when MQTT had to reconnect
mqtt_client.on('reconnect', function() {
    console.log("event => Trying to reconnect...");
});

// Listen to messages on the MQTT bus
mqtt_client.on('message', function(topic, message) {

    var topic_message = message.toString().trim();

    console.log('event => MQTT_MESSAGE_RECEIVED, topic: "' + topic + '", message: ' + topic_message);

    // Service Adapter Subscribe topic
    // IMPORTANT: should publish to flaneur/mqtt_service_adapter/add_topic and the message should be the new topic.
    // New topics are appended automatically to subscriptions.txt
    if (topic === 'flaneur/mqtt_service_adapter/add_topic') {
        fs.readFile(subscriptions_txt, 'utf8', function (error, data) {
            if (error) {
                console.log("event => ERROR: ", error);
            }
            // Checks topic_message exists on subscriptions.txt
            if (data.indexOf(topic_message) <= 0){
                // Subscribe to new topic
                mqtt_client.subscribe(topic_message);
                // Append new topic into subscriptions.txt
                fs.appendFile(subscriptions_txt, topic_message + '\n', function (error) {
                    if (error) {
                        return console.log("event => ERROR: ", error);
                    }
                    console.log("event => Topic subscribed: " + topic_message);
                });
            } else {
                console.log("event => ERROR: Topic already subscribed");
            }
        });
    }

    // Service Adapter Unsubscribe topic
    // IMPORTANT: should publish to flaneur/mqtt_service_adapter/remove_topic and the message should be the topic to be removed.
    // Topics are removed automatically in subscriptions.txt
    if (topic === 'flaneur/mqtt_service_adapter/remove_topic') {
        // Unsubscribe from MQTT and removes topic in subscriptions.txt
        fs.readFile(subscriptions_txt, 'utf8', function (error, data) {
            if (error) {
                console.log("event => ERROR: ", error);
            }
            // Checks topic_message exists on subscriptions.txt
            if (data.indexOf(topic_message) >= 0){
                // Unsubscribe to topic
                mqtt_client.unsubscribe(topic_message);
                // removes topic from subscriptions.txt file
                var fileData = data.toString();
                fileData = fileData.replace(topic_message, "");
                // removes empty lines from subscriptions.txt file
                fileData = fileData.replace(/^\s*[\r\n]/gm, "");
                fs.writeFile(subscriptions_txt, fileData, 'utf8', function (error) {
                    if (error) {
                        console.log("event => ERROR: ", error);
                    }
                    console.log("event => Topic unsubscribed: " + topic_message);
                });
            } else {
                console.log("event => ERROR: Topic does not exists");
            }
        });
    }

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
        mqtt_client.subscribe(line);
        console.log("event => Topic subscribed: " + line);
    }
});

// Sends test after 1 second (1000ms)
setTimeout(function(){
    var namespace = 'flaneur/tusd/upload_success';
    mqtt_client.publish(namespace, 'test ok');
}, 1000);
