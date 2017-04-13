package main

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/jingqiuELE/tinyvpn/internal/encrypt"
	"github.com/jingqiuELE/tinyvpn/internal/rsautil"
	"github.com/jingqiuELE/tinyvpn/internal/session"
)

const BUFFERSIZE = 1500

func authGetSession(serverAddr string, port int, keyfile string) (sk session.Index, ss session.Secret, ip net.IP, err error) {
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

	secret, err := session.NewSecret()
	if err != nil {
		log.Error("Failed to create new Secret:", err)
		return
	}
	log.Debug("Generated session secret: ", secret)
	ss = *secret

	rsautil.LoadKey(keyfile, &publicKey)
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
	data, err := encrypt.Decrypt((*secret)[:], iv, payload)

	copy(sk[:], data[0:session.IndexLen])

	index := session.IndexLen
	ip = net.IPv4(data[index], data[index+1], data[index+2], data[index+3])
	log.Info("Assigned IP address:", ip.String())

	return
}
