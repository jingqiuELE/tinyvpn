package main

import (
	"fmt"
	"net"
	"session"
	"strconv"
	"sync"
)

type AuthServer struct {
	sync.RWMutex
	secretMap map[session.Key]session.Secret
}

func newAuthServer(serverIP string, port int) (*AuthServer, error) {
	m := make(map[session.Key]session.Secret)
	a := &AuthServer{secretMap: m}

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

	go func() {
		for {
			conn, err := l.AcceptTCP()
			if err != nil {
				fmt.Println("Error: ", err)
				continue
			}
			go a.handleAuthConn(conn)
		}
	}()
	return a, err
}

func (a *AuthServer) handleAuthConn(conn *net.TCPConn) error {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Failed to read:", err)
		return err
	}

	sk, err := session.NewKey()
	if err != nil {
		fmt.Println("Failed to create new Key:", err)
		return err
	}

	secret := []byte{23, 42, 17, 5}

	a.Lock()
	a.secretMap[*sk] = secret
	a.Unlock()
	_, err = conn.Write(sk[:])
	if err != nil {
		fmt.Println("Failed to write Session Key:", err)
		return err
	}
	return err

}
