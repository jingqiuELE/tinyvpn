package main

import (
	"sync"

	sh "github.com/codeskyblue/go-sh"
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
)

type Book struct {
	sync.RWMutex
	ipToSession map[string]Index
	sessionToIp map[Index]string
}

type BookServer struct {
	book *Book
	pOut chan *Packet
	pIn  chan *Packet
	tun  *water.Interface
}

func newBookServer(pOut, pIn chan *Packet, vpnnet string, tun *water.Interface) (bs *BookServer, err error) {
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

		bs.book.Add(src_ip.String(), p.Sk)
		_, err := bs.tun.Write(p.Data)
		if err != nil {
			log.Error("Error writing to tun!", err)
		}
	}
}

/* Handle traffic from target to client */
func (bs *BookServer) listenTun() error {
	buffer := make([]byte, MTU)
	for {
		n, err := bs.tun.Read(buffer)
		if err != nil {
			log.Error("Error reading from tunnel.")
			return err
		}

		dst_ip := waterutil.IPv4Destination(buffer[:n])

		p := new(Packet)
		sk := bs.book.getSession(dst_ip.String())
		p.Sk = sk
		p.Data = buffer[:n]
		log.Debug("Book: to client", p)
		bs.pOut <- p
	}
}

func newBook() *Book {
	b := new(Book)
	b.ipToSession = make(map[string]Index)
	b.sessionToIp = make(map[Index]string)
	return b
}
func (b *Book) getSession(ip string) Index {
	b.RLock()
	key := b.ipToSession[ip]
	b.RUnlock()
	return key
}

func (b *Book) getIp(sessionIndex Index) string {
	b.RLock()
	ip := b.sessionToIp[sessionIndex]
	b.RUnlock()
	return ip
}

func (b *Book) Add(ip string, sessionIndex Index) {
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
