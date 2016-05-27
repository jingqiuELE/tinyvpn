package main

import (
	"fmt"
	"net"
	"packet"
	"strconv"
)

func startConnection(serverAddr string, port int, eOut, eIn chan packet.Packet) error {
	connServer := serverAddr + ":" + strconv.Itoa(port)
	raddr, err := net.ResolveUDPAddr("udp", connServer)
	if err != nil {
		fmt.Println("Error resolving connServer:", err)
		return err
	}

	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("Error resolving localaddr:", err)
		return err
	}

	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		fmt.Println("Error dail remote:", err)
		return err
	}
	defer conn.Close()

	go handleOut(conn, eOut)
	go handleIn(conn, eIn)

	return err
}

func handleOut(conn *net.UDPConn, eOut chan packet.Packet) {
	for {
		p := <-eOut
		buf := packet.MarshalToSlice(p)
		_, err := conn.Write(buf)
		if err != nil {
			fmt.Println("Error writing to Connection:", err)
		}
	}
}

func handleIn(conn *net.UDPConn, eIn chan packet.Packet) {
	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Failed to read from Connection:", err)
			continue
		}

		p, err := packet.UnmarshalSlice(buf[:n])
		if err != nil {
			fmt.Println("Failed to unmarshal data:", err)
			continue
		}
		eIn <- p
	}
}
