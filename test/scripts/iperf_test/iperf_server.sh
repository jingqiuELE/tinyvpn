#!/bin/bash

display_help () {
    echo "iperf_server.sh [PROTOCOL]"
    echo "PROTOCOL: udp or tcp"
}

if [ $# -ne 1 ]; then
    echo "Please follow bellow format:"
    display_help
    exit 1
fi

PROTO=$1

# Limit iperf buffer length to be the mss of UDP packet, which equals to 
# MSS = MTU - Header_size = 1412 - 28 = 1384
if [ $PROTO == "udp" ]; then
    PROTO="-u -l 1384"
else
    PROTO=""
fi

cd ./scripts
iperf $PROTO -s > iperf_server.log &
