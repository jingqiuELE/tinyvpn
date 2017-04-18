#!/bin/bash
./petrel server -k "./keys/private.key" -s 10.0.3.100 &
iptables -t nat -A POSTROUTING -o server-eth0 -j MASQUERADE
