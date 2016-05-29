package main

import (
	"errors"
	"fmt"
	"github.com/songgao/water"
	"os"
	"os/signal"
	"packet"
	"syscall"

	flag "github.com/spf13/pflag"
)

func main() {

	const channelSize = 10
	var (
		tcpPort, udpPort, authPort int
		serverAddr, vpnnet         string

		encryptedOutChan = make(chan packet.Packet, channelSize)
		plainOutChan     = make(chan packet.Packet, channelSize)
		encryptedInChan  = make(chan packet.Packet, channelSize)
		plainInChan      = make(chan packet.Packet, channelSize)
	)

	flag.IntVarP(&authPort, "auth", "a", 7282, "Port for the authentication service to listen to.")
	flag.IntVarP(&tcpPort, "tcp", "t", 8272, "TCP port to listen to, 0 to disable tcp")
	flag.IntVarP(&udpPort, "udp", "u", 8272, "UDP port to listen to, 0 to disable udp")
	flag.StringVarP(&serverAddr, "serverAddr", "s", "0.0.0.0", "IP address the server suppose to listen to, e.g. 127.0.0.1")
	flag.StringVarP(&vpnnet, "vpnnet", "n", "10.82.72.0/24", "Subnet netmask for the VPN subnet, e.g. 10.0.0.1/24")
	flag.Parse()

	fmt.Printf("Values of the config are: Auth %v, TCP %v, UDP %v, ServerAddr %v, Vpnnet %v\n", authPort, tcpPort, udpPort, serverAddr, vpnnet)

	_, err := newAuthServer(serverAddr, authPort)
	if err != nil {
		fmt.Printf("Failed to create AuthServer %v\n", err)
		return
	}

	_, err = newConnServer(serverAddr, tcpPort, udpPort, encryptedOutChan, encryptedInChan)
	if err != nil {
		fmt.Println("Failed to create ConnServer:", err)
		return
	}

	//To be passed with a (Type *AuthServer), so that EncryptServer can access session secret.
	_, err = newEncryptServer(encryptedOutChan, encryptedInChan, plainOutChan, plainInChan)
	if err != nil {
		fmt.Println("Failed to create EncryptServer:", err)
		return
	}

	tun, err := water.NewTUN("")
	if err != nil {
		fmt.Println("Error creating tun interface", err)
		return
	}

	b, err := newBookServer(plainOutChan, plainInChan, vpnnet, tun)
	if err != nil {
		fmt.Println("Failed to create BookServer:", err)
		return
	}

	//Receive system signal to stop the server.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

		s := <-c
		fmt.Println("Received signal", errors.New(s.String()))
	}()

	b.start()
}
