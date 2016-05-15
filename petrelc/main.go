package main

import (
	"errors"
	"fmt"
	"github.com/songgao/water"
	flag "github.com/spf13/pflag"
	"io"
	"net"
	"os"
	"os/signal"
	"packet"
	"syscall"
)

func main() {
	var carrierProtocol, authServer, connServer string

	flag.StringVar(&carrierProtocol, "carrierProtocol", "udp", "protocol for the underlying layer of tiny vpn.")
	flag.StringVar(&authServer, "authServer", "0.0.0.0:7282", "authServer's address")
	flag.StringVar(&connServer, "connServer", "0.0.0.0:8272", "connServer's address")
	flag.Parse()

	sk, err := contactAuthServer(authServer)
	if err != nil {
		fmt.Println("Failed to auth myself:", err)
		return
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

		s := <-c
		fmt.Println("Received signal", errors.New(s.String()))
	}()

	tun, err := water.NewTUN("")
	if err != nil {
		fmt.Printf("Error is %v\n", err)
		return
	}

	startClientSession(connServer, sk, tun)

	fmt.Println("process quit")

	return
}

func contactAuthServer(authServer string) (sk [6]byte, err error) {
	raddr, err := net.ResolveTCPAddr("tcp", authServer)
	if err != nil {
		fmt.Println("Error when resolving authServer:", err)
		return
	}

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		fmt.Println("Dail error:", err)
		return
	}
	defer conn.Close()

	buf := []byte("Hello from client")
	_, err = conn.Write(buf)
	if err != nil {
		fmt.Println("Error writing:", err)
		return
	}

	_, err = conn.Read(sk[:])
	if err != nil {
		fmt.Println("Read error:", err)
		if err != io.EOF {
			return
		}
	}
	return
}

func startClientSession(connServer string, sk [6]byte, tun *water.Interface) (err error) {
	raddr, err := net.ResolveUDPAddr("udp", connServer)
	if err != nil {
		fmt.Println("Error resolving connServer:", err)
		return
	}
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("Error resolving localaddr:", err)
		return
	}

	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		fmt.Println("Error dail remote:", err)
		return
	}
	defer conn.Close()

	c := make(chan []byte)
	go startListenTun(tun, c)
	for {
		data := <-c
		p := packet.NewPacket(sk, data)
		buf, err := packet.Marshal(p)
		if err != nil {
			fmt.Println("Failed to unmarshal the Packet:", err)
			continue
		}
		conn.Write(buf)
	}
	return
}

func startListenTun(tun *water.Interface, c chan []byte) (err error) {
	buffer := make([]byte, 1500)
	for {
		_, err = tun.Read(buffer)
		if err != nil {
			fmt.Println("Error reading from tunnel:", err)
			return
		}
		c <- buffer
	}
}
