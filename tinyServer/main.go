package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var carrierProtocol, secret, serverAddr, vpnnet string

	flag.StringVar(&carrierProtocol, "carrierProtocol", "udp", "protocol for the underlying layer of tiny vpn.")
	flag.StringVar(&secret, "secret", "", "secret")
	flag.StringVar(&serverAddr, "serverAddr", "0.0.0.0:9525", "Server's listening address")
	flag.StringVar(&vpnnet, "vpnnet", "10.0.200.1/24", "vpn's internal network gateway and netmask.")
	flag.Parse()

	tinyServer, err := CreateTinyServer(secret, carrierProtocol, serverAddr, vpnnet)
	defer tinyServer.Close()

	err = tinyServer.Run()
	if err != nil {
		fmt.Println("Error running tinyServer.")
	}
	fmt.Println("tinyvpn server started")

	//Receive system signal to stop the server.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

		s := <-c
		fmt.Println("Received signal", errors.New(s.String()))
	}()

	fmt.Println("process quit")

	return
}
