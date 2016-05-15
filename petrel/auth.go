package main

import (
	"fmt"
	"net"
	"strconv"
)

func startAuthenticationServer(serverIP string, port int, s *SessionMap) (err error) {
	serverAddr := serverIP + ":" + strconv.Itoa(port)
	listenAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error when resoving TCP Address!")
		return err
	}
	l, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		return
	}
	defer l.Close()
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Println("Error: ", err)
		} else {
			go handleAuthConn(conn, s)
		}
	}
	return
}

func handleAuthConn(conn *net.TCPConn, s *SessionMap) {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Failed to read:", err)
		return
	}
	sk, err := NewSessionKey()
	if err != nil {
		fmt.Println("Failed to create new SessionKey:", err)
		return
	}
	s.Update(*sk, TConnection{})
	_, err = conn.Write(sk[:])
	if err != nil {
		fmt.Println("Failed to write SessionKey:", err)
		return
	}

}
