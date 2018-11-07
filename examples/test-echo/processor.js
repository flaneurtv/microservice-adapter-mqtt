var readline = require('readline');
var fs = require('fs');
var expandenv = require('expandenv');

// Reads JSON from stdin
var rl_stdin = readline.createInterface({
    input: process.stdin
});
// When line comes from stdin
rl_stdin.on('line', function(line){
    try {
        // Assings ENV variables to an object
        obj = process.env;
        if (line !== null) {
            // first we check if the payload is valid JSON
            json_payload = JSON.parse(line);
            obj.PAYLOAD = line.trim().toString();
            // Generates datetime
            obj.CREATED_AT = new Date().toISOString();
            // Open JSON template
            var rl_template = readline.createInterface({
                input: fs.createReadStream('./service-processor/schema-test-echo.json')
            });
            var result = '';
            // concat lines from template
            rl_template.on('line', function(line) {
                if (line !== null) {
                    // Trims JSON and then output a single line string
                    result += line.trim().toString();
                }
            });
            // All lines read and concatenated
            rl_template.on('close', () => {
                // Replace ENV variables in the JSON template with values and print to stdout
                console.log(expandenv(result));
            });
        }
    } catch (e) {
        console.log('{"topic": "' + process.env.NAMESPACE + '/log", "message": "http-request: error"}');
    }
});
