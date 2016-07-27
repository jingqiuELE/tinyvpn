package main

import (
	"errors"
	"github.com/op/go-logging"
	"github.com/songgao/water"
	"logger"
	"os"
	"os/signal"
	"packet"
	"syscall"
	"tunnel"

	flag "github.com/spf13/pflag"
)

var log = logger.Get(logging.DEBUG)

func main() {

	const channelSize = 10
	var (
		tcpPort, udpPort, authPort int
		serverAddr, vpnnet         string

		eOut = make(chan packet.Packet, channelSize)
		pOut = make(chan packet.Packet, channelSize)
		eIn  = make(chan packet.Packet, channelSize)
		pIn  = make(chan packet.Packet, channelSize)
	)

	flag.IntVarP(&authPort, "auth", "a", 7282, "Port for the authentication service to listen to.")
	flag.IntVarP(&tcpPort, "tcp", "t", 8272, "TCP port to listen to, 0 to disable tcp")
	flag.IntVarP(&udpPort, "udp", "u", 8272, "UDP port to listen to, 0 to disable udp")
	flag.StringVarP(&serverAddr, "serverAddr", "s", "0.0.0.0", "IP address the server suppose to listen to, e.g. 127.0.0.1")
	flag.StringVarP(&vpnnet, "vpnnet", "n", "172.0.1.1/24", "Subnet netmask for the VPN subnet, e.g. 172.0.0.1/24")
	flag.Parse()

	log.Infof("Values of the config are: Auth %v, TCP %v, UDP %v, ServerAddr %v, vpnnet %v", authPort, tcpPort, udpPort, serverAddr, vpnnet)

	a, err := newAuthServer(serverAddr, authPort, vpnnet)
	if err != nil {
		log.Errorf("Failed to create AuthServer %v", err)
		return
	}

	_, err = newConnServer(serverAddr, tcpPort, udpPort, eOut, eIn)
	if err != nil {
		log.Error("Failed to create ConnServer:", err)
		return
	}

	_, err = newEncryptServer(a, eOut, eIn, pOut, pIn)
	if err != nil {
		log.Error("Failed to create EncryptServer:", err)
		return
	}

	tun, err := water.NewTUN("")
	if err != nil {
		log.Error("Error creating tun interface", err)
		return
	}
	err = tunnel.AddAddr(tun, vpnnet)
	if err != nil {
		log.Error(err)
		return
	}
	err = tunnel.Bringup(tun)
	if err != nil {
		log.Error(err)
		return
	}
	err = SetNAT()
	if err != nil {
		log.Error(err)
		return
	}

	b, err := newBookServer(pOut, pIn, vpnnet, tun)
	if err != nil {
		log.Error("Failed to create BookServer:", err)
		return
	}

	//Receive system signal to stop the server.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

		s := <-c
		log.Notice("Received signal", errors.New(s.String()))
	}()

	b.start()
}
