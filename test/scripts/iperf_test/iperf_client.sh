#!/bin/bash

display_help () {
    echo "iperf_client.sh [PROTOCOL]"
    echo "PROTOCOL: udp or tcp"
}

if [ $# -ne 1 ]; then
    echo "Please follow bellow format:"
    display_help
    exit 1
fi

PROTO=$1

# Limit iperf buffer length to be the mss of UDP packet (MTU-Header_size)
if [ $PROTO == "udp" ]; then
    PROTO="-u -l 1432"
else
    PROTO=""
fi

cd ./scripts
iperf $PROTO -c 10.0.5.100 > iperf_client.log &
