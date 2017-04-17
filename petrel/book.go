package main

import (
	"fmt"
	"net"
	"packet"
	"session"
	"sync"

	"github.com/codeskyblue/go-sh"
	"github.com/jingqiuELE/tinyvpn/internal/packet"
	"github.com/jingqiuELE/tinyvpn/internal/session"
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
)

type Book struct {
	sync.RWMutex
	ipToSession map[string]session.Index
	sessionToIp map[session.Index]string
}

type BookServer struct {
	book *Book
	pOut chan packet.Packet
	pIn  chan packet.Packet
	tun  *water.Interface
}

func newBookServer(pOut, pIn chan packet.Packet, vpnnet string, tun *water.Interface) (bs *BookServer, err error) {
	bs = &BookServer{
		book: newBook(),
		pOut: pOut,
		pIn:  pIn,
		tun:  tun,
	}
	return
}

func (bs *BookServer) start() {
	go bs.listenTun()
	for {
		p, ok := <-bs.pIn
		if !ok {
			log.Error("Failed to read from pIn:")
			return
		}
		src_ip := waterutil.IPv4Source(p.Data)
		log.Debug("Book: from client", src_ip)
		log.Debug("Book: datalen=", len(p.Data))

		bs.book.Add(src_ip.String(), p.Header.Sk)
		_, err := bs.tun.Write(p.Data)
		if err != nil {
			log.Error("Error writing to tun!", err)
		}
	}
}

/* Handle traffic from target to client */
func (bs *BookServer) listenTun() error {
	buffer := make([]byte, packet.MTU)
	for {
		n, err := bs.tun.Read(buffer)
		if err != nil {
			log.Error("Error reading from tunnel.")
			return err
		}

		dst_ip := waterutil.IPv4Destination(buffer[:n])
		log.Debug("Book: to client", dst_ip)

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
	b.ipToSession = make(map[string]session.Index)
	b.sessionToIp = make(map[session.Index]string)
	return b
}
func (b *Book) getSession(ip string) session.Index {
	b.RLock()
	key := b.ipToSession[ip]
	b.RUnlock()
	return key
}

func (b *Book) getIp(sessionIndex session.Index) string {
	b.RLock()
	ip := b.sessionToIp[sessionIndex]
	b.RUnlock()
	return ip
}

func (b *Book) Add(ip string, sessionIndex session.Index) {
	b.Lock()
	b.ipToSession[ip] = sessionIndex
	b.sessionToIp[sessionIndex] = ip
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
		log.Info("Default route interface is", default_if)
		err = sh.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", default_if, "-j", "MASQUERADE").Run()
	}
	return err
}
