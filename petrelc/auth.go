package main

import (
	"fmt"
	"io"
	"net"
	"session"
	"strconv"
)

const BUFFERSIZE = 1500

func authGetSession(serverAddr string, port int) (session.SessionKey, error) {
	var sk session.SessionKey
	authServer := serverAddr + ":" + strconv.Itoa(port)
	raddr, err := net.ResolveTCPAddr("tcp", authServer)
	if err != nil {
		fmt.Println("Error when resolving authServer:", err)
		return sk, err
	}

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		fmt.Println("Dail error:", err)
		return sk, err
	}
	defer conn.Close()

	secret := []byte("Hello world")
	_, err = conn.Write(secret)
	if err != nil {
		fmt.Println("Error writing:", err)
		return sk, err
	}

	buf := make([]byte, BUFFERSIZE)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Println("Read error:", err)
		if err != io.EOF {
			return sk, err
		}
	}

	copy(sk[:], buf[:6])
	return sk, err
}
