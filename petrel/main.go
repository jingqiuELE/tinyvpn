package main

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/op/go-logging"
	"github.com/songgao/water"

	flag "github.com/spf13/pflag"
)

var log = GetLogger(logging.DEBUG)

func main() {
	const channelSize = 10
	var (
		eOut = make(chan *Packet, channelSize)
		pOut = make(chan *Packet, channelSize)
		eIn  = make(chan *Packet, channelSize)
		pIn  = make(chan *Packet, channelSize)
	)

	authPort := flag.IntP("auth", "a", 7282, "Port for the authentication service to listen to.")
	tcpPort := flag.IntP("tcp", "t", 8272, "TCP port to listen to, 0 to disable tcp")
	udpPort := flag.IntP("udp", "u", 8272, "UDP port to listen to, 0 to disable udp")
	serverAddr := flag.StringP("serverAddr", "s", "0.0.0.0", "IP address the server suppose to listen to, e.g. 127.0.0.1")
	vpnnet := flag.StringP("vpnnet", "n", "172.0.1.1/24", "Subnet netmask for the VPN subnet, e.g. 172.0.0.1/24")
	mode := os.Args[1]

	switch mode {
	case "client":
		keyfile := flag.StringP("keyfile", "k", "./public.key", "Public key file used for client authentication")
		protocol := flag.StringP("protocol", "t", "tcp", "Protocol used for connection, tcp or udp")
		port := flag.IntP("port", "p", 8272, "Port to connect to")
		runClient(*serverAddr, *authPort, *port, *keyfile, *protocol, eOut, pOut, eIn, eIn)
	case "server":
		keyfile := flag.StringP("keyfile", "k", "./private.key", "Private key for server to use for authentication")
		runServer(*serverAddr, *vpnnet, *keyfile, *authPort, *tcpPort, *udpPort, eOut, pOut, eIn, pIn)
	default:
		log.Fatalf("Petrel needs to run either as client or server, please try\n\npetrel client\n")
	}

}

func runClient(serverAddr string, authPort, port int, keyfile string, protocol string, eOut, pOut, eIn, pIn chan *Packet) {

	sk, secret, ip, err := authGetSession(serverAddr, authPort, keyfile)
	if err != nil {
		log.Error("Failed to auth myself:", err)
		return
	}

	err = startListenTun(pIn, pOut, ip)
	if err != nil {
		log.Error("Failed to start Tunnel Listener:", err)
		return
	}

	err = startEncrypt(eOut, eIn, pOut, pIn, sk, secret)
	if err != nil {
		log.Error("Failed to create EncryptServer:", err)
		return
	}

	err = startConnection(serverAddr, protocol, port, eOut, eIn)
	if err != nil {
		log.Error("Faild to create Connection:", err)
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Kill)
	signal.Notify(sigs, syscall.SIGTERM)
	s := <-sigs
	log.Notice("Received signal", errors.New(s.String()))
	log.Notice("process quit")
}

// TODO: Continue from here
func runServer(serverAddr, vpnnet, keyfile string, authPort, tcpPort, udpPort int, eOut, pOut, eIn, pIn chan *Packet) {
	a, err := newAuthServer(serverAddr, authPort, vpnnet, keyfile)
	if err != nil {
		log.Errorf("Failed to create AuthServer %v", err)
		return
	}

	_, err = newConnServer(serverAddr, tcpPort, udpPort, eOut, eIn)
	if err != nil {
		log.Error("Failed to create ConnServer:", err)
		return
	}

	_, err = newEncryptServer(a, eOut, eIn, pOut, pIn)
	if err != nil {
		log.Error("Failed to create EncryptServer:", err)
		return
	}

	tun, err := water.NewTUN("")
	if err != nil {
		log.Error("Error creating tun interface", err)
		return
	}
	err = AddAddr(tun, vpnnet)
	if err != nil {
		log.Error(err)
		return
	}
	err = SetMtu(tun, MTU)
	if err != nil {
		log.Error(err)
		return
	}
	err = Bringup(tun)
	if err != nil {
		log.Error(err)
		return
	}
	err = SetNAT()
	if err != nil {
		log.Error(err)
		return
	}

	b, err := newBookServer(pOut, pIn, vpnnet, tun)
	if err != nil {
		log.Error("Failed to create BookServer:", err)
		return
	}

	//Receive system signal to stop the server.
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Kill)
		signal.Notify(sigs, syscall.SIGTERM)
		s := <-sigs
		log.Notice("Received signal", errors.New(s.String()))
	}()

	b.start()

}
