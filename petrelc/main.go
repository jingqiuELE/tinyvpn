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
		encryptedOutChan           = make(chan packet.Packet, channelSize)
		plainOutChan               = make(chan packet.Packet, channelSize)
		encryptedInChan            = make(chan packet.Packet, channelSize)
		plainInChan                = make(chan packet.Packet, channelSize)
	)

	flag.StringVarP(&serverAddr, "serverAddr", "s", "0.0.0.0", "IP address of the server")
	flag.IntVarP(&authPort, "auth", "a", 7282, "Port for the authentication service.")
	flag.IntVarP(&tcpPort, "tcp", "t", 8272, "TCP port of connServer")
	flag.IntVarP(&udpPort, "udp", "u", 8272, "UDP port of connServer")
	flag.Parse()

	sk, err := authGetSession(serverAddr, authPort)
	if err != nil {
		fmt.Println("Failed to auth myself:", err)
		return
	}

	err = startEncrypt(encryptedOutChan, encryptedInChan,
		plainOutChan, plainInChan, sk)
	if err != nil {
		fmt.Println("Failed to create EncryptServer:", err)
		return
	}

	err = startConnection(serverAddr, udpPort, encryptedOutChan, encryptedInChan)
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
