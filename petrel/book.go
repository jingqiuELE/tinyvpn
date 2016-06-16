package main

import (
	"fmt"
	"net"
	"packet"
	"session"
	"sync"

	"github.com/codeskyblue/go-sh"
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
)

type Book struct {
	sync.RWMutex
	ipToSession map[string]session.Key
	sessionToIp map[session.Key]string
}

type BookServer struct {
	book *Book
	pOut chan packet.Packet
	pIn  chan packet.Packet
	tun  *water.Interface
}

func newBookServer(pOut, pIn chan packet.Packet, vpnnet string, tun *water.Interface) (bs *BookServer, err error) {
	bs = new(BookServer)
	bs.book = newBook()
	bs.pOut = pOut
	bs.pIn = pIn
	bs.tun = tun
	return
}

func (bs *BookServer) start() {

	go bs.listenTun()

	for {
		p, ok := <-bs.pOut
		if !ok {
			log.Error("Failed to read from pOut:")
			return
		}
		src_ip := waterutil.IPv4Source(p.Data)

		sk := new(session.Key)
		copy(sk[:], p.Header.Sk[:session.KeyLen])
		bs.book.Add(src_ip.String(), *sk)
		bs.tun.Write(p.Data)
	}
}

const BUFFERSIZE = 1500

/* Handle internet traffic for the vpnnet */
func (bs *BookServer) listenTun() error {
	buffer := make([]byte, BUFFERSIZE)
	for {
		_, err := bs.tun.Read(buffer)
		if err != nil {
			log.Error("Error reading from tunnel.")
			return err
		}
		p := packet.NewPacket()
		p.SetData(buffer)

		// Get the IP address and find the corresponding session
		var ip net.IP
		if waterutil.IsIPv4(buffer) {
			ip = waterutil.IPv4Destination(buffer)
		} else if waterutil.IsIPv6(buffer) {
			// Water does not handle IPv6 packet destination yet or IPv6 works totally different?
			fmt.Errorf("IPv6 packet cannot be properly handled atm.")
		} else {
			panic("Packet not of IPv4 nor IPv6")
		}

		sk := bs.book.getSession(ip.String())
		p.Header.Sk = sk[:]

		bs.pIn <- *p
	}
}

func newBook() *Book {
	b := new(Book)
	b.ipToSession = make(map[string]session.Key)
	b.sessionToIp = make(map[session.Key]string)
	return b
}
func (b *Book) getSession(ip string) session.Key {
	b.RLock()
	key := b.ipToSession[ip]
	b.RUnlock()
	return key
}

func (b *Book) getIp(sessionKey session.Key) string {
	b.RLock()
	ip := b.sessionToIp[sessionKey]
	b.RUnlock()
	return ip
}

func (b *Book) Add(ip string, sessionKey session.Key) {
	b.Lock()
	b.ipToSession[ip] = sessionKey
	b.sessionToIp[sessionKey] = ip
	b.Unlock()
}

// shell scripts to manipulate tun network interface device.

func getDefaultRouteIf() (string, error) {
	ifce, err := sh.Command("ip", "route", "show", "default").Command("awk", "/default/ {print $5}").Output()
	return string(ifce[:]), err
}

func SetNAT() error {
	var err error
	default_if, err := getDefaultRouteIf()
	if err != nil {
		log.Error("Cannot get the default routing interface")
	} else {
		log.Error("Default route interface is", default_if)
		err = sh.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", default_if, "-j", "MASQUERADE").Run()
	}
	return err
}
