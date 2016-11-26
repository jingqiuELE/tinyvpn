#!/bin/bash

display_help () {
    echo "snc.sh [PROTOCOL]"
    echo "PROTOCOL: udp or tcp"
}

if [ $# -ne 1 ]; then
    echo "Please follow bellow format:"
    display_help
    exit 1
fi

PROTO=$1

if [ $PROTO == "udp" ]; then
    PROTO="-u"
else
    PROTO=""
fi

cd ./scripts
nc $PROTO -lvp 2222 > random.out &
