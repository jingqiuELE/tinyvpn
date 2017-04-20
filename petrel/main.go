package main

import (
	"os"

	"github.com/op/go-logging"

	flag "github.com/spf13/pflag"
)

var log = GetLogger(logging.DEBUG)

func main() {
	const channelSize = 10

	authPort := flag.IntP("auth", "a", 7282, "Port for the authentication service to listen to.")
	port := flag.IntP("port", "p", 8272, "Port to connect to")
	serverAddr := flag.StringP("serverAddr", "s", "0.0.0.0", "IP address the server suppose to listen to, e.g. 127.0.0.1")
	vpnnet := flag.StringP("vpnnet", "n", "172.0.1.1/24", "Subnet netmask for the VPN subnet, e.g. 172.0.0.1/24")
	mode := os.Args[1]

	switch mode {
	case "client":
		keyfile := flag.StringP("keyfile", "k", "./public.key", "Public key file used for client authentication")
		protocol := flag.StringP("protocol", "t", "tcp", "Protocol used for connection, tcp or udp")
		flag.Parse()

		err := runClient(*serverAddr, *authPort, *port, *keyfile, *protocol)
		if err != nil {
			log.Fatal(err)
		}
	case "server":
		keyfile := flag.StringP("keyfile", "k", "./private.key", "Private key for server to use for authentication")
		flag.Parse()

		err := runServer(*serverAddr, *vpnnet, *keyfile, *authPort, *port, *port)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("Petrel needs to run either as client or server, please try\n\npetrel client\n")
	}

	select {} // Block to prevent exit
}

func runClient(serverAddr string, authPort, port int, keyfile string, protocol string) error {

	// Authenticate with auth server first
	sk, secret, ip, err := authGetSession(serverAddr, authPort, keyfile)
	if err != nil {
		log.Error("Failed to auth myself:", err)
		return err
	}

	// Create local tun device
	book := &StaticBook{sk, ip.String()}
	toTun, fromTun, err := startTUN(ip.String(), MTU, book)
	if err != nil {
		log.Error("Failed to start Tunnel device:", err)
		return err
	}

	toServer, fromServer, err := startConnection(serverAddr, protocol, port)
	if err != nil {
		log.Error("Failed to start connection to server:", err)
		return err
	}

	ss := &StaticSecretSource{secret}

	go encryptPackets(fromTun, toServer, ss)
	go decryptPackets(fromServer, toTun, ss)

	return nil
}

func runServer(serverAddr, vpnnet, keyfile string, authPort, tcpPort, udpPort int) error {
	a, err := newAuthServer(serverAddr, authPort, vpnnet, keyfile)
	if err != nil {
		log.Errorf("Failed to create AuthServer %v", err)
		return err
	}

	// Create local tun device
	book := newDynBook()
	toTun, fromTun, err := startTUN(vpnnet, MTU, book)
	if err != nil {
		log.Error("Failed to start Tunnel device:", err)
		return err
	}

	toClient, fromClient, err := startConnServer(serverAddr, tcpPort, udpPort)
	if err != nil {
		log.Error("Failed to create ConnServer:", err)
		return err
	}

	err = SetNAT()
	if err != nil {
		log.Error("Failed to setup NAT", err)
		return err
	}

	go encryptPackets(fromTun, toClient, a)
	go decryptPackets(fromClient, toTun, a)

	return nil
}
