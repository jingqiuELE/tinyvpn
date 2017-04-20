package main

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
)

var ErrUser = errors.New("User crenditial is wrong")
var ErrIPAddrPoolFull = errors.New("IPAddrPool is full")
var privateKey rsa.PrivateKey

type AuthServer struct {
	sync.RWMutex
	secretMap  map[Index]Secret
	ipAddrPool IPAddrPool
}

const BUFFERSIZE = 1500

type SecretSource interface {
	getSecret(sk Index) (Secret, bool)
}

type StaticSecretSource struct {
	Secret Secret
}

func (s *StaticSecretSource) getSecret(sk Index) (Secret, bool) {
	return s.Secret, true
}

func newAuthServer(serverIP string, port int, vpnnet string, keyfile string) (*AuthServer, error) {
	m := make(map[Index]Secret)
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
	a.ipAddrPool = NewIPAddrPool(internalIP, ipNet)

	LoadKey(keyfile, &privateKey)

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

func (a *AuthServer) handleAuthConn(conn *net.TCPConn) error {
	var session_secret Secret

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Error("Failed to read:", err)
		return err
	}

	secret := GetRandomSessionKey()
	err = rsa.DecryptPKCS1v15SessionKey(rand.Reader, &privateKey, buf[:n],
		secret)

	if err != nil {
		fmt.Println("Error decrypting session key ", err.Error())
		os.Exit(1)
	}
	log.Debug("Decrypted session key is ", secret)

	sk, err := NewIndex()
	if err != nil {
		log.Error("Failed to create new Index:", err)
		return err
	}
	log.Debug("session index:", sk)
	log.Debug("Received session secret: ", secret[:SecretLen])

	copy(session_secret[:], secret[:SecretLen])
	a.setSecret(*sk, session_secret)

	ip, err := a.ipAddrPool.Get()
	if err != nil {
		return err
	}
	assignIP := ip.IP.To4()
	log.Debug("Assigning IP: ", assignIP)

	buf = append((*sk)[:], assignIP...)
	encrypt_data, iv, err := Encrypt(secret, aes.BlockSize, buf)
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

func (a *AuthServer) setSecret(sk Index, secret Secret) {
	a.Lock()
	a.secretMap[sk] = secret
	a.Unlock()
}

func (a *AuthServer) getSecret(sk Index) (Secret, bool) {
	a.RLock()
	secret, ok := a.secretMap[sk]
	a.RUnlock()
	return secret, ok
}

func authGetSession(serverAddr string, port int, keyfile string) (sk Index, ss Secret, ip net.IP, err error) {
	var publicKey rsa.PublicKey

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

	secret, err := NewSecret()
	if err != nil {
		log.Error("Failed to create new Secret:", err)
		return
	}
	log.Debug("Generated session secret: ", secret)
	ss = *secret

	LoadKey(keyfile, &publicKey)
	encrypt_session_key, err := rsa.EncryptPKCS1v15(rand.Reader, &publicKey, (*secret)[:])
	if err != nil {
		fmt.Println("Error encrypting session key ", err.Error())
		return
	}

	_, err = conn.Write(encrypt_session_key)
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

	iv := buf[0:aes.BlockSize]
	payload := buf[aes.BlockSize:n]
	data, err := Decrypt((*secret)[:], iv, payload)

	copy(sk[:], data[0:IndexLen])

	index := IndexLen
	ip = net.IPv4(data[index], data[index+1], data[index+2], data[index+3])
	log.Info("Assigned IP address:", ip.String())

	return
}
