package main

import (
	"errors"
	"fmt"
	"github.com/songgao/water"
	flag "github.com/spf13/pflag"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var carrierProtocol, authServer, connServer string

	flag.StringVarP(&carrierProtocol, "carrierProtocol", "p", "udp", "protocol for the underlying layer of tiny vpn.")
	flag.StringVarP(&authServer, "authServer", "a", "0.0.0.0:7282", "authServer's address")
	flag.StringVarP(&connServer, "connServer", "c", "0.0.0.0:8272", "connServer's address")
	flag.Parse()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

		s := <-c
		fmt.Println("Received signal", errors.New(s.String()))
	}()

	sk, err := getSession(authServer)

	if err != nil {
		fmt.Println("Failed to auth myself:", err)
		return
	}

	tun, err := water.NewTUN("")
	if err != nil {
		fmt.Printf("Error is %v\n", err)
		return
	}

	startConnection(connServer, sk, tun)

	fmt.Println("process quit")

	return
}
