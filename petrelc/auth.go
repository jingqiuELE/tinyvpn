package main

import (
	"fmt"
	"io"
	"net"
	"session"
	"strconv"
)

const BUFFERSIZE = 1500

func authGetSession(serverAddr string, port int) (sk session.Index, ss session.Secret, ip net.IP, err error) {
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
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Read error:", err)
		if err != io.EOF {
			return
		}
	}

	if n >= (session.IndexLen + session.SecretLen + net.IPv4len) {
		prev := 0
		next := session.IndexLen
		copy(sk[:], buf[prev:next])

		prev = next
		next += session.SecretLen
		copy(ss[:], buf[prev:next])

		prev = next
		ip = net.IPv4(buf[prev], buf[prev+1],
			buf[prev+2], buf[prev+3])
	}

	log.Info("Assigned IP address:", ip.String())

	return
}
