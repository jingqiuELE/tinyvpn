package main

import (
	"os/exec"
	"strings"
	"sync"

	"github.com/songgao/water"
)

type Book interface {
	getSession(ip string) (Index, bool)
	getIp(sessionIndex Index) (string, bool)
	Add(ip string, sessionIndex Index)
}

type StaticBook struct {
	Sk Index
	Ip string
}

func (b *StaticBook) getSession(ip string) (Index, bool) {
	return b.Sk, true
}

func (b *StaticBook) getIp(sessionIndex Index) (string, bool) {
	return b.Ip, true
}

func (b *StaticBook) Add(ip string, sessionIndex Index) {}

type DynBook struct {
	sync.RWMutex
	ipToSession map[string]Index
	sessionToIp map[Index]string
}

type BookServer struct {
	book *DynBook
	pOut chan *Packet
	pIn  chan *Packet
	tun  *water.Interface
}

//func newBookServer(pOut, pIn chan *Packet, vpnnet string, tun *water.Interface) (bs *BookServer, err error) {
//bs = &BookServer{
//book: newDynBook(),
//pOut: pOut,
//pIn:  pIn,
//tun:  tun,
//}
//return
//}

//func (bs *BookServer) start() {
//go bs.listenTun()
//for {
//p, ok := <-bs.pIn
//if !ok {
//log.Error("Failed to read from pIn:")
//return
//}
//src_ip := waterutil.IPv4Source(p.Data)
//log.Debug("Book: from client", src_ip)
//log.Debug("Book: datalen=", len(p.Data))

//bs.book.Add(src_ip.String(), p.Sk)
//_, err := bs.tun.Write(p.Data)
//if err != nil {
//log.Error("Error writing to tun!", err)
//}
//}
//}

//[> Handle traffic from target to client <]
//func (bs *BookServer) listenTun() error {
//buffer := make([]byte, MTU)
//for {
//n, err := bs.tun.Read(buffer)
//if err != nil {
//log.Error("Error reading from tunnel.")
//return err
//}

//dst_ip := waterutil.IPv4Destination(buffer[:n])

//p := new(Packet)
//sk, _ := bs.book.getSession(dst_ip.String())
//p.Sk = sk
//p.Data = buffer[:n]
//log.Debug("Book: to client", p)
//bs.pOut <- p
//}
//}

func newDynBook() *DynBook {
	b := new(DynBook)
	b.ipToSession = make(map[string]Index)
	b.sessionToIp = make(map[Index]string)
	return b
}

func (b *DynBook) getSession(ip string) (Index, bool) {
	b.RLock()
	sk, ok := b.ipToSession[ip]
	b.RUnlock()
	return sk, ok
}

func (b *DynBook) getIp(sessionIndex Index) (string, bool) {
	b.RLock()
	ip, ok := b.sessionToIp[sessionIndex]
	b.RUnlock()
	return ip, ok
}

func (b *DynBook) Add(ip string, sessionIndex Index) {
	b.Lock()
	b.ipToSession[ip] = sessionIndex
	b.sessionToIp[sessionIndex] = ip
	b.Unlock()
}

func SetNAT() error {
	route, err := exec.Command("ip", "route", "show", "default", "0.0.0.0/0").Output()
	if err != nil {
		log.Error("Cannot get the default routing interface")
		return err
	}
	parts := strings.Split(string(route), " ")
	ifce := parts[4] // default via 192.168.1.1 dev wlp3s0  proto static  metric 600

	log.Info("Default route interface is", ifce)
	_, err = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", ifce, "-j", "MASQUERADE").Output()
	return err
}
