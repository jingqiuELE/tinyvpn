package main

import (
	"fmt"
	"net"
	"strconv"

	"github.com/songgao/water"
	flag "github.com/spf13/pflag"
)

type SessionKey [6]byte
type Packet struct {
	iv         [8]byte
	sessionKey SessionKey
	length     uint16
	data       []byte
}

type Connection interface {
	writePacket(p Packet) error
}

type Session struct {
	conn   Connection
	secret []byte
}

type IpSessionMap struct {
	ipToSession map[string]SessionKey
	sessionToIp map[SessionKey]string
}

func (m *IpSessionMap) getSession(ip string) SessionKey {
	return m.ipToSession[ip]
}

func (m *IpSessionMap) getIp(sessionKey SessionKey) string {
	return m.sessionToIp[sessionKey]
}

func (m *IpSessionMap) Add(ip string, sessionKey SessionKey) {
	m.ipToSession[ip] = sessionKey
	m.sessionToIp[sessionKey] = ip
}

func main() {

	const channelSize = 10
	var (
		tcpPort, udpPort, authPort int
		serverAddr, netmask        string
		tun                        water.Interface

		encryptedOutChan = make(chan Packet, channelSize)
		plainOutChan     = make(chan Packet, channelSize)
		encryptedInChan  = make(chan Packet, channelSize)
		plainInChan      = make(chan Packet, channelSize)

		sessionMap   = make(map[SessionKey]Session)
		ipSessionMap IpSessionMap
	)

	flag.IntVarP(&authPort, "auth", "a", 7282, "Port for the authentication service to listen to.")
	flag.IntVarP(&tcpPort, "tcp", "t", 8272, "TCP port to listen to, 0 to disable tcp")
	flag.IntVarP(&udpPort, "udp", "u", 8272, "UDP port to listen to, 0 to disable udp")
	flag.StringVarP(&serverAddr, "serverAddr", "s", "0.0.0.0", "IP address the server suppose to listen to, e.g. 127.0.0.1")
	flag.StringVarP(&netmask, "netmask", "n", "10.82.72.0/24", "Subnet netmask for the VPN subnet, e.g. 10.0.0.1/24")
	flag.Parse()

	fmt.Printf("Values of the config are: Auth %v, TCP %v, UDP %v, ServerAddr %v, Netmask %v\n", authPort, tcpPort, udpPort, serverAddr, netmask)

	tun, err := createTunInterface()
	if err != nil {
		fmt.Printf("Error is %v\n", err)
		return
	}

	err = startAuthenticationServer(serverAddr, authPort, sessionMap)
	if err != nil {
		fmt.Printf("Error is %v\n", err)
		return
	}

	if tcpPort != 0 {
		err = startTCPListener(serverAddr, tcpPort, encryptedOutChan, sessionMap)
		if err != nil {
			fmt.Printf("Error is %v\n", err)
			return
		}
	}
	if udpPort != 0 {
		startUDPListener(serverAddr, udpPort, encryptedOutChan, sessionMap)
		if err != nil {
			fmt.Printf("Error is %v\n", err)
			return
		}
	}

	startPacketReturner(encryptedInChan, sessionMap)
	startPacketDecrypter(encryptedOutChan, plainOutChan, sessionMap)

	startPacketEncrypter(encryptedInChan, plainInChan, sessionMap)
	startTunPacketSink(plainOutChan, tun, ipSessionMap)
	startTunListener(plainInChan, tun, ipSessionMap)
}

func startAuthenticationServer(serverAddr string, port int, sessionMap map[SessionKey]Session) (err error) {
	l, err := net.Listen("tcp", serverAddr+":"+strconv.Itoa(port))
	if err != nil {
		return
	}
	defer l.Close()
	return
}

func createTunInterface() (ifce water.Interface, err error) {
	return
}

// Listen to TCP socket and put the packets to the encrypted out chan.
func startTCPListener(serverAddr string, port int, encryptedOutChan chan Packet, sessionMap map[SessionKey]Session) (err error) {
	l, err := net.Listen("tcp", serverAddr+":"+strconv.Itoa(port))
	if err != nil {
		return
	}
	defer l.Close()
	// TODO: Read the packets and put them into the encryptedOutChan.
	return
}

// Listen to UDP socket and put the packets to the encrypted out chan.
func startUDPListener(serverAddr string, port int, encryptedOutChan chan Packet, sessionMap map[SessionKey]Session) (err error) {
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr+":"+strconv.Itoa(port))
	if err != nil {
		return
	}
	l, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return
	}
	defer l.Close()
	// TODO: Read the packets and put them into the encryptedOutChan.
	return
}

func startPacketReturner(encryptedInChan chan Packet, sessionMap map[SessionKey]Session) (err error) {

	return
}

func startPacketDecrypter(encryptedOutChan, plainOutChan chan Packet, sessionMap map[SessionKey]Session) (err error) {

	return
}

func startPacketEncrypter(encryptedInChan, plainInChan chan Packet, sessionMap map[SessionKey]Session) (err error) {

	return
}

func startTunPacketSink(plainOutChan chan Packet, ifce water.Interface, ipSessionMap IpSessionMap) (err error) {

	return
}

func startTunListener(plainInChan chan Packet, ifce water.Interface, ipSessionMap IpSessionMap) (err error) {

	return
}
