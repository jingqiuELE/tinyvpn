package main

import (
	"fmt"
	"github.com/codeskyblue/go-sh"
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
	"packet"
	"session"
	"tunnel"
)

type Book struct {
	ipToSession map[string]session.Key
	sessionToIp map[session.Key]string
}

type BookServer struct {
	book   *Book
	pOut   chan packet.Packet
	pIn    chan packet.Packet
	vpnnet string
	tun    *water.Interface
}

func newBookServer(pOut, pIn chan packet.Packet, vpnnet string) (*BookServer, error) {
	var err error
	bs := new(BookServer)
	bs.book = newBook()
	bs.pOut = pOut
	bs.pIn = pIn
	bs.vpnnet = vpnnet
	bs.tun, err = water.NewTUN("")
	if err != nil {
		fmt.Println("Error creating tun interface", err)
		return bs, err
	}

	return bs, err
}

func (bs *BookServer) start() {
	err := tunnel.AddAddr(bs.tun, bs.vpnnet)
	if err != nil {
		return
	}

	err = tunnel.Bringup(bs.tun)
	if err != nil {
		return
	}

	err = SetNAT()
	if err != nil {
		return
	}

	go bs.listenTun()

	for {
		p, ok := <-bs.pOut
		if !ok {
			fmt.Println("Failed to read from pOut:")
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
			fmt.Println("Error reading from tunnel.")
			return err
		}
		p := packet.NewPacket()
		p.SetData(buffer)
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
	return b.ipToSession[ip]
}

func (b *Book) getIp(sessionKey session.Key) string {
	return b.sessionToIp[sessionKey]
}

func (b *Book) Add(ip string, sessionKey session.Key) {
	b.ipToSession[ip] = sessionKey
	b.sessionToIp[sessionKey] = ip
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
		fmt.Println("Cannot get the default routing interface")
	} else {
		fmt.Println("Default route interface is", default_if)
		err = sh.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", default_if, "-j", "MASQUERADE").Run()
	}
	return err
}
