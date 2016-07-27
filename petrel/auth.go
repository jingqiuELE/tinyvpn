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
	secretMap  map[session.Index]session.Secret
	ipAddrPool ippool.IPAddrPool
}

func newAuthServer(serverIP string, port int, vpnnet string) (*AuthServer, error) {
	userMap = map[string]string{
		"apple":  "juice",
		"banana": "shake",
		"orange": "raw",
	}

	m := make(map[session.Index]session.Secret)
	a := &AuthServer{secretMap: m}

	serverAddr := serverIP + ":" + strconv.Itoa(port)
	listenAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		log.Error("Resoving TCP Address:", err)
		return a, err
	}

	l, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		return a, err
	}

	internalIP, ipNet, err := net.ParseCIDR(vpnnet)
	if err != nil {
		log.Errorf("Failed to parse CIDR: %s: %v", err, vpnnet)
		return a, err
	}
	a.ipAddrPool = ippool.NewIPAddrPool(internalIP, ipNet)

	go func() {
		for {
			conn, err := l.AcceptTCP()
			if err != nil {
				log.Error(err)
				continue
			}
			go a.handleAuthConn(conn)
		}
	}()
	return a, err
}

func dumpString(s string) {
	fmt.Println("dump start. len=", len(s))
	fmt.Printf("% x ", s)
}

func (a *AuthServer) handleAuthConn(conn *net.TCPConn) error {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Error("Failed to read:", err)
		return err
	}

	s := string(buf[:n])
	data := strings.Split(s, ":")
	user := data[0]
	rp := data[1]
	p, ok := userMap[user]
	if !ok || (p != rp) {
		log.Error("User/Password incorrect!", ok)
		return ErrUser
	}

	sk, err := session.NewIndex()
	if err != nil {
		log.Error("Failed to create new Index:", err)
		return err
	}
	log.Debug("session key:", sk)

	secret, err := session.NewSecret()
	if err != nil {
		log.Error("Failed to create new Secret:", err)
		return err
	}

	a.setSecret(*sk, *secret)

	ip, err := a.ipAddrPool.Get()
	if err != nil {
		return err
	}
	assignIP := ip.IP.To4()

	/* buf should hold sk, secret and assigned ip address. */
	buf = append((*sk)[:], (*secret)[:]...)
	buf = append(buf[:], assignIP...)
	_, err = conn.Write(buf)
	if err != nil {
		log.Error("Failed to sendback response:", err)
		return err
	}
	return err
}

func (a *AuthServer) setSecret(sk session.Index, secret session.Secret) {
	a.Lock()
	a.secretMap[sk] = secret
	a.Unlock()
}

func (a *AuthServer) getSecret(sk session.Index) (session.Secret, bool) {
	a.RLock()
	secret, ok := a.secretMap[sk]
	a.RUnlock()
	return secret, ok
}
