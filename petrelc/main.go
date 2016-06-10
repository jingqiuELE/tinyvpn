package main

import (
	"errors"
	"fmt"
	flag "github.com/spf13/pflag"
	"os"
	"os/signal"
	"packet"
	"syscall"
)

func main() {
	const channelSize = 10

	var (
		tcpPort, udpPort, authPort int
		serverAddr                 string
		eOut                       = make(chan packet.Packet, channelSize)
		pOut                       = make(chan packet.Packet, channelSize)
		eIn                        = make(chan packet.Packet, channelSize)
		pIn                        = make(chan packet.Packet, channelSize)
	)

	flag.StringVarP(&serverAddr, "serverAddr", "s", "0.0.0.0", "IP address of the server")
	flag.IntVarP(&authPort, "auth", "a", 7282, "Port for the authentication service.")
	flag.IntVarP(&tcpPort, "tcp", "t", 8272, "TCP port of connServer")
	flag.IntVarP(&udpPort, "udp", "u", 8272, "UDP port of connServer")
	flag.Parse()

	sk, ip, err := authGetSession(serverAddr, authPort)
	if err != nil {
		fmt.Println("Failed to auth myself:", err)
		return
	}

	err = startListenTun(pIn, pOut, ip)
	if err != nil {
		fmt.Println("Failed to start Tunnel Listener:", err)
		return
	}

	err = startEncrypt(eOut, eIn, pOut, pIn, sk)
	if err != nil {
		fmt.Println("Failed to create EncryptServer:", err)
		return
	}

	err = startConnection(serverAddr, udpPort, eOut, eIn)
	if err != nil {
		fmt.Println("Faild to create Connection:", err)
		return
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
		s := <-c
		fmt.Println("Received signal", errors.New(s.String()))
	}()

	fmt.Println("process quit")

	return
}
