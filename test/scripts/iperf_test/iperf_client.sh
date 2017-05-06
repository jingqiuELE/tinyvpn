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
BW=""

# Limit iperf buffer length to be the mss of UDP packet (MTU-Header_size)
# By default, iperf will limit the bandwidth of UDP traffic to 1Mbps. Set the limit to be 1Gbps instead.

if [ $PROTO == "udp" ]; then
    PROTO="-u -l 1384"
    BW="-b 1g"
else
    PROTO=""
fi

cd ./scripts
iperf $PROTO -c 10.0.5.100 $BW > iperf_client.log &
