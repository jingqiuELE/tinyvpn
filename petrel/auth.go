package main

import (
	"errors"
	"fmt"
	"ippool"
	"net"
	"session"
	"strconv"
	"strings"
	"sync"
)

var userMap map[string]string

var ErrUser = errors.New("User crenditial is wrong")
var ErrIPAddrPoolFull = errors.New("IPAddrPool is full")

type AuthServer struct {
	sync.RWMutex
	secretMap  map[session.Key]session.Secret
	ipAddrPool ippool.IPAddrPool
}

func newAuthServer(serverIP string, port int, vpnnet string) (*AuthServer, error) {
	userMap = map[string]string{
		"apple":  "juice",
		"banana": "shake",
		"orange": "raw",
	}

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

	internalIP, ipNet, err := net.ParseCIDR(vpnnet)
	if err != nil {
		fmt.Println("Error in vpnnet format: %V", vpnnet)
		return a, err
	}
	a.ipAddrPool = ippool.NewIPAddrPool(internalIP, ipNet)

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

func dumpString(s string) {
	fmt.Println("dump start. len=", len(s))
	for i := 0; i < len(s); i++ {
		fmt.Printf("%x ", s[i])
	}
	fmt.Println("dump complete")
}

func (a *AuthServer) handleAuthConn(conn *net.TCPConn) error {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Failed to read:", err)
		return err
	}

	s := string(buf[:n])
	fmt.Println("Received: ", s)
	data := strings.Split(s, ":")
	user := data[0]
	rp := data[1]
	p, ok := userMap[user]
	if !ok || (p != rp) {
		fmt.Println("User/Password incorrect!", ok)
		return ErrUser
	}

	sk, err := session.NewKey()
	if err != nil {
		fmt.Println("Failed to create new Key:", err)
		return err
	}

	secret, err := session.NewSecret()
	if err != nil {
		fmt.Println("Failed to create new Secret:", err)
		return err
	}

	a.Lock()
	a.secretMap[*sk] = *secret
	a.Unlock()

	ip, err := a.ipAddrPool.Get()
	if err != nil {
		return err
	}
	assignIP := ip.IP.To4()

	buf = append((*sk)[:], assignIP...)
	_, err = conn.Write(buf)
	if err != nil {
		fmt.Println("Failed to sendback response:", err)
		return err
	}
	return err
}
