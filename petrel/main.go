package main

import (
	"fmt"
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
	flag "github.com/spf13/pflag"
	"packet"
)

func main() {

	const channelSize = 10
	var (
		tcpPort, udpPort, authPort int
		serverAddr, netmask        string
		tun                        *water.Interface

		encryptedOutChan = make(chan packet.Packet, channelSize)
		plainOutChan     = make(chan packet.Packet, channelSize)
		encryptedInChan  = make(chan packet.Packet, channelSize)
		plainInChan      = make(chan packet.Packet, channelSize)

		s            *SessionMap
		ipSessionMap IpSessionMap
	)

	flag.IntVarP(&authPort, "auth", "a", 7282, "Port for the authentication service to listen to.")
	flag.IntVarP(&tcpPort, "tcp", "t", 8272, "TCP port to listen to, 0 to disable tcp")
	flag.IntVarP(&udpPort, "udp", "u", 8272, "UDP port to listen to, 0 to disable udp")
	flag.StringVarP(&serverAddr, "serverAddr", "s", "0.0.0.0", "IP address the server suppose to listen to, e.g. 127.0.0.1")
	flag.StringVarP(&netmask, "netmask", "n", "10.82.72.0/24", "Subnet netmask for the VPN subnet, e.g. 10.0.0.1/24")
	flag.Parse()

	fmt.Printf("Values of the config are: Auth %v, TCP %v, UDP %v, ServerAddr %v, Netmask %v\n", authPort, tcpPort, udpPort, serverAddr, netmask)

	tun, err := water.NewTUN("")
	if err != nil {
		fmt.Printf("Error is %v\n", err)
		return
	}

	s = NewSessionMap()

	err = startAuthenticationServer(serverAddr, authPort, s)
	if err != nil {
		fmt.Printf("Error is %v\n", err)
		return
	}

	if tcpPort != 0 {
		err = startTCPListener(serverAddr, tcpPort, encryptedOutChan, s)
		if err != nil {
			fmt.Printf("Error is %v\n", err)
			return
		}
	}
	if udpPort != 0 {
		startUDPListener(serverAddr, udpPort, encryptedOutChan, s)
		if err != nil {
			fmt.Printf("Error is %v\n", err)
			return
		}
	}

	startPacketReturner(encryptedInChan, s)
	startPacketDecrypter(encryptedOutChan, plainOutChan, s)

	startPacketEncrypter(encryptedInChan, plainInChan, s)

	startTunPacketSink(plainOutChan, tun, ipSessionMap)
	startTunListener(plainInChan, tun, ipSessionMap)
}

func startPacketReturner(encryptedInChan chan packet.Packet, s *SessionMap) (err error) {
	return
}

func startPacketDecrypter(encryptedOutChan, plainOutChan chan packet.Packet, s *SessionMap) (err error) {

	return
}

func startPacketEncrypter(encryptedInChan, plainInChan chan packet.Packet, s *SessionMap) (err error) {

	return
}

func startTunPacketSink(plainOutChan chan packet.Packet, ifce *water.Interface, ipSessionMap IpSessionMap) (err error) {

	return
}

func startTunListener(plainInChan chan packet.Packet, ifce *water.Interface, ipSessionMap IpSessionMap) {
	const bufferSize = 65535
	go func() {
		for {
			buffer := make([]byte, bufferSize)
			_, err := ifce.Read(buffer)
			if err != nil {
				fmt.Println("Error reading from tunnel.")
			}
			var dest net.IP
			if waterutil.IsIPv4(buffer) {
				dest = waterutil.IPv4Source(buffer)
			}
			if waterutil.IsIPv6(buffer) {
				// TODO: IPv6 support
			}

			if dest != nil {
				destStr := dest.String()
				sessionKey := ipSessionMap.getSession(destStr)
				if sessionKey != nil {
					p := Packet{
						sessionKey: sessionKey,
					}
					plainInChan <- p
				}
			}
			// TODO: Create packet from buffer
			//plainInChan <- buffer
		}
	}()
	return
}
