#!/bin/bash
cd /projects/tinyvpn
./myvpn-server -secret milk -logtostderr -v 3 -up-script ./if-up.sh &
iptables -t nat -A POSTROUTING -o server-eth0 -j MASQUERADE
