#!/bin/bash

MININET_NAME="tinyvpn_mininet"
docker run --rm --net=container:$MININET_NAME docker-tcpdump -l -i s2-eth1 --immediate-mode -w - | wireshark -k -i - &
docker run --rm --net=container:$MININET_NAME docker-tcpdump -l -i s3-eth1 --immediate-mode -w - | wireshark -k -i - &

