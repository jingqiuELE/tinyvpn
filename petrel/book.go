package main

import (
	"github.com/codeskyblue/go-sh"
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
	"packet"
	"session"
	"sync"
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
		p, ok := <-bs.pIn
		if !ok {
			log.Error("Failed to read from pIn:")
			return
		}
		src_ip := waterutil.IPv4Source(p.Data)
		log.Debug("client ip is", src_ip)

		bs.book.Add(src_ip.String(), p.Header.Sk)
		bs.tun.Write(p.Data)
	}
}

const BUFFERSIZE = 1500

/* Handle traffic from target to client */
func (bs *BookServer) listenTun() error {
	buffer := make([]byte, BUFFERSIZE)
	for {
		_, err := bs.tun.Read(buffer)
		if err != nil {
			log.Error("Error reading from tunnel.")
			return err
		}
		dst_ip := waterutil.IPv4Destination(buffer)
		sk := bs.book.getSession(dst_ip.String())
		p := packet.NewPacket()
		p.Header.Sk = sk
		p.SetData(buffer)
		bs.pOut <- *p
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
		log.Info("Default route interface is", default_if)
		err = sh.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", default_if, "-j", "MASQUERADE").Run()
	}
	return err
}
