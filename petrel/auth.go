package main

import (
	"fmt"
	"net"
	"session"
	"strconv"
)

type AuthServer struct {
	secretMap map[session.SessionKey]session.Secret
}

func newAuthServer(serverIP string, port int) (*AuthServer, error) {
	a := new(AuthServer)
	a.secretMap = make(map[session.SessionKey]session.Secret)

	serverAddr := serverIP + ":" + strconv.Itoa(port)
	listenAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error when resoving TCP Address!")
		return a, err
	}

	l, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		return a, err
	}
	defer l.Close()

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}
		go a.handleAuthConn(conn)
	}
	return a, err
}

func (a *AuthServer) handleAuthConn(conn *net.TCPConn) error {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Failed to read:", err)
		return err
	}

	sk, err := session.NewSessionKey()
	if err != nil {
		fmt.Println("Failed to create new SessionKey:", err)
		return err
	}

	secret := []byte{23, 42, 17, 5}
	a.secretMap[*sk] = secret
	_, err = conn.Write(sk[:])
	if err != nil {
		fmt.Println("Failed to write SessionKey:", err)
		return err
	}
	return err

}
