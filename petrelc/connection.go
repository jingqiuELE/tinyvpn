package main

import (
	"fmt"
	"github.com/songgao/water"
	"net"
	"packet"
	"strconv"
)

func startUDPConnection(serverAddr string, port int, sk [6]byte, tun *water.Interface) (err error) {
	connServer := serverAddr + ":" + strconv.Itoa(port)
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
		p := packet.NewPacket(data)
		p.Header.Sk = sk
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
