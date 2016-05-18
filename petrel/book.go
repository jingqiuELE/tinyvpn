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
	ipToSession map[string]session.SessionKey
	sessionToIp map[session.SessionKey]string
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
	err := tunnel.Bringup(bs.tun, bs.vpnnet)
	if err != nil {
		return
	}

	err = SetNAT()
	if err != nil {
		return
	}

	go bs.listenTun()

	for {
		p, ok := <-bs.pIn
		if !ok {
			fmt.Println("Failed to read from pIn:")
			return
		}
		src_ip := waterutil.IPv4Source(p.Data)
		bs.book.Add(src_ip.String(), p.Header.Sk)
		bs.tun.Write(p.Data)
	}
}

func (bs *BookServer) listenTun() error {
	buffer := make([]byte, BUFFERSIZE)
	for {
		_, err := bs.tun.Read(buffer)
		if err != nil {
			fmt.Println("Error reading from tunnel.")
			return err
		}
		p := packet.NewPacket(buffer)
		bs.pOut <- *p
	}
}

func newBook() *Book {
	b := new(Book)
	b.ipToSession = make(map[string]session.SessionKey)
	b.sessionToIp = make(map[session.SessionKey]string)
	return b
}
func (b *Book) getSession(ip string) session.SessionKey {
	return b.ipToSession[ip]
}

func (b *Book) getIp(sessionKey session.SessionKey) string {
	return b.sessionToIp[sessionKey]
}

func (b *Book) Add(ip string, sessionKey session.SessionKey) {
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
