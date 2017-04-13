package main

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/jingqiuELE/tinyvpn/internal/logger"
	"github.com/jingqiuELE/tinyvpn/internal/packet"
	"github.com/op/go-logging"
	flag "github.com/spf13/pflag"
)

var log = logger.Get(logging.DEBUG)

func main() {
	const channelSize = 10

	var (
		authPort, connPort       int
		serverAddr, connProtocol string
		keyfile                  string
		eOut                     = make(chan packet.Packet, channelSize)
		pOut                     = make(chan packet.Packet, channelSize)
		eIn                      = make(chan packet.Packet, channelSize)
		pIn                      = make(chan packet.Packet, channelSize)
	)

	flag.StringVarP(&serverAddr, "serverAddr", "s", "0.0.0.0", "IP address of server")
	flag.IntVarP(&authPort, "authPort", "a", 7282, "Port of authServer.")
	flag.IntVarP(&connPort, "connPort", "c", 8272, "port of connServer")
	flag.StringVarP(&connProtocol, "connProtocol", "p", "udp", "transport protocol to connServer")
	flag.StringVarP(&keyfile, "keyfile", "k", "./public.key", "public key for the Auth")
	flag.Parse()

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

	err = startConnection(serverAddr, connProtocol, connPort, eOut, eIn)
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

	return
}
