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
	var (
		tcpPort, udpPort, authPort int
		serverAddr                 string
	)

	flag.StringVarP(&serverAddr, "serverAddr", "s", "0.0.0.0", "IP address of the server")
	flag.IntVarP(&authPort, "auth", "a", 7282, "Port for the authentication service.")
	flag.IntVarP(&tcpPort, "tcp", "t", 8272, "TCP port of connServer")
	flag.IntVarP(&udpPort, "udp", "u", 8272, "UDP port of connServer")
	flag.Parse()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

		s := <-c
		fmt.Println("Received signal", errors.New(s.String()))
	}()

	sk, err := getSession(serverAddr, authPort)

	if err != nil {
		fmt.Println("Failed to auth myself:", err)
		return
	}

	tun, err := water.NewTUN("")
	if err != nil {
		fmt.Printf("Error is %v\n", err)
		return
	}

	startUDPConnection(serverAddr, udpPort, sk, tun)

	fmt.Println("process quit")

	return
}
