#!/usr/bin/env python

import sys, json, datetime, string, os, pytz

def time_format(dt):
    return "%s:%.3f%s" % (
        dt.strftime('%Y-%m-%dT%H:%M'),
        float("%.3f" % (dt.second + dt.microsecond / 1e6)),
        dt.strftime('%z')
    )

while True:
    try:
        # Assings ENV variables to an object
        obj = os.environ
        # Reads JSON from stdin
        obj["PAYLOAD"] = sys.stdin.readline().replace("\n", "")

        if obj["PAYLOAD"] != "":
            # first we check if the payload is valid JSON
            json_payload = json.loads(obj["PAYLOAD"])
            # Generates datetime ISO 8601 format ex. 2017-04-24T21:16:26.678Z
            obj["CREATED_AT"] = time_format(datetime.datetime.now(pytz.utc)).replace('+0000', 'Z')
            # Open JSON template
            f = open('./service-processor/schema-test-echo.json', 'r')
            lines = f.readlines()
            # Strips JSON and then output a single line string
            json_str = ''.join([line.strip() for line in lines])
            # Replace ENV variables in the JSON template with values and print to stdout
            # Must use print and not sys.stdout.write otherwise it will not work
            print(string.Template(json_str).substitute(obj))
            sys.stdout.flush()

    except ValueError:
        # If JSON from stdin is invalid
        sys.stdout.write('{"topic": "' + obj["NAMESPACE"] + '/log", "message": "http-request: error"}' + '\n')
        sys.stdout.flush()
