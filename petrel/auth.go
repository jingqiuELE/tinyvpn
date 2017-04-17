package main

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/jingqiuELE/tinyvpn/internal/encrypt"
	"github.com/jingqiuELE/tinyvpn/internal/ippool"
	"github.com/jingqiuELE/tinyvpn/internal/rsautil"
	"github.com/jingqiuELE/tinyvpn/internal/session"
)

var ErrUser = errors.New("User crenditial is wrong")
var ErrIPAddrPoolFull = errors.New("IPAddrPool is full")
var privateKey rsa.PrivateKey

type AuthServer struct {
	sync.RWMutex
	secretMap  map[session.Index]session.Secret
	ipAddrPool ippool.IPAddrPool
}

func newAuthServer(serverIP string, port int, vpnnet string, keyfile string) (*AuthServer, error) {
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

	rsautil.LoadKey(keyfile, &privateKey)

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
	var session_secret session.Secret

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Error("Failed to read:", err)
		return err
	}

	secret := rsautil.GetRandomSessionKey()
	err = rsa.DecryptPKCS1v15SessionKey(rand.Reader, &privateKey, buf[:n],
		secret)

	if err != nil {
		fmt.Println("Error decrypting session key ", err.Error())
		os.Exit(1)
	}
	log.Debug("Decrypted session key is ", secret)

	sk, err := session.NewIndex()
	if err != nil {
		log.Error("Failed to create new Index:", err)
		return err
	}
	log.Debug("session index:", sk)
	log.Debug("Received session secret: ", secret[:session.SecretLen])

	copy(session_secret[:], secret[:session.SecretLen])
	a.setSecret(*sk, session_secret)

	ip, err := a.ipAddrPool.Get()
	if err != nil {
		return err
	}
	assignIP := ip.IP.To4()
	log.Debug("Assigning IP: ", assignIP)

	buf = append((*sk)[:], assignIP...)
	encrypt_data, iv, err := encrypt.Encrypt(secret, aes.BlockSize, buf)
	if err != nil {
		return err
	}
	/* buf should hold sk, secret and assigned ip address. */
	buf = append(iv, encrypt_data[:]...)
	_, err = conn.Write(buf)
	if err != nil {
		log.Error("Failed to sendback response:", err)
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
