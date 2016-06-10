package main

import (
	"fmt"
	"io"
	"net"
	"session"
	"strconv"
)

const BUFFERSIZE = 1500

func authGetSession(serverAddr string, port int) (sk session.Key, ip net.IP, err error) {
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

	secret := []byte("apple:juice")
	_, err = conn.Write(secret)
	if err != nil {
		fmt.Println("Error writing:", err)
		return
	}

	buf := make([]byte, BUFFERSIZE)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Println("Read error:", err)
		if err != io.EOF {
			return
		}
	}

	copy(sk[:], buf[:6])
	copy(ip[:], buf[6:])

	return
}
