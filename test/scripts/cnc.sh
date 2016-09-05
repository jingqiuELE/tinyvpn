#!/bin/bash

display_help () {
    echo "cnc [COUNT] [PROTOCOL] [TARGET] [PORT]"
    echo "COUNT: number of bytes to send"
    echo "PROTOCOL: udp or tcp"
    echo "TARGET: destination address"
    echo "PORT: destination port"
}

if [ $# -ne 4 ]; then
    echo "Please follow bellow format:"
    display_help
    exit 1
fi

COUNT=$1
PROTO=$2
TARGET=$3
PORT=$4

if [ $PROTO == "udp" ]; then
    PROTO="-u"
else
    PROTO=""
fi

cat /dev/urandom | base64 | head -c $COUNT | nc -q 1 $PROTO $TARGET $PORT
