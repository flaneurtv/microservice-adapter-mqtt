#!/bin/bash

# This is a demo processor which shall be overwritten by your own logic.
# This demo listens for tick messages and sends out a tick_reply on each one.
read stdin_msg;

# assigns JSON topic data to variable topic using jq
topic=$( jq -r  '.topic' <<< "${stdin_msg}" )

# if topic is "flaneur/tusd/upload_success" then send to stdout upload_status
if [ $topic == "flaneur/tusd/upload_success" ]; then
    topic="flaneur/tusd/upload_status"
    created_at="672534868293"

    printf '{"topic": "%s", "created_at": "%s"}\n' "$topic" "$created_at"
fi
