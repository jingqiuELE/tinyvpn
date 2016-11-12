#!/bin/bash

display_help () {
    echo "cnc [COUNT] [PROTOCOL]"
    echo "COUNT: number of bytes to send"
    echo "PROTOCOL: udp or tcp"
}

if [ $# -ne 2 ]; then
    echo "Please follow bellow format:"
    display_help
    exit 1
fi

COUNT=$1
PROTO=$2

if [ $PROTO == "udp" ]; then
    PROTO="-u"
else
    PROTO=""
fi

cd /projects/tinyvpn/scripts
cat /dev/urandom | base64 | head -c $COUNT > random.in 
nc -q 1 $PROTO 10.0.5.100 2222 < random.in
