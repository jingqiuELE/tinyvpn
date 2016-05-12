#!/bin/bash
cd /projects/tinyvpn
./myvpn-client -secret milk -server-addr 10.0.3.100:9525 -logtostderr -v 3 -up-script ./if-up.sh &
ip route add 10.0.3.100 via 10.0.1.1 dev client-eth0 src 10.0.1.100
ip route del default
sleep 2
ip route add default dev tun0
