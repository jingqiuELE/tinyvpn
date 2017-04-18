package main

import (
	"net"

	"github.com/songgao/water"
)

/* Handle client's traffic, wrap each packet with outer IP Header */
func startListenTun(pIn, pOut chan *Packet, ip net.IP) error {
	tun, err := water.NewTUN("")
	if err != nil {
		log.Error(err)
		return err
	}

	err = AddAddr(tun, ip.String())
	if err != nil {
		return err
	}

	err = SetMtu(tun, MTU)
	if err != nil {
		return err
	}

	err = Bringup(tun)
	if err != nil {
		return err
	}

	go handleTunOut(tun, pOut)
	go handleTunIn(tun, pIn)

	return err
}

/* handle traffic from client to target */
func handleTunOut(tun *water.Interface, pOut chan *Packet) {
	for {
		p, err := Decode(tun)
		if err != nil {
			log.Error("Reading from tunnel:", err)
			return
		}

		pOut <- p
	}
}

/* handle traffic from target to client. */
func handleTunIn(tun *water.Interface, pIn chan *Packet) {
	for {
		p := <-pIn
		tun.Write(p.Data)
	}
}
