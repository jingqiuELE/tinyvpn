package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
)

func authGetSession(serverAddr string, port int) (sk [6]byte, err error) {
	authServer := serverAddr + ":" + strconv.Itoa(port)
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
