package main

import (
	"github.com/songgao/water"
	"net"
	"packet"
	"tunnel"
)

/* Handle client's traffic, wrap each packet with outer IP Header */
func startListenTun(pIn, pOut chan packet.Packet, ip net.IP) error {
	tun, err := water.NewTUN("")
	if err != nil {
		log.Error(err)
		return err
	}

	err = tunnel.AddAddr(tun, ip.String())
	if err != nil {
		return err
	}

	err = tunnel.Bringup(tun)
	if err != nil {
		return err
	}

	go handleTunOut(tun, pOut)
	go handleTunIn(tun, pIn)

	return err
}

/* handle traffic from client to target */
func handleTunOut(tun *water.Interface, pOut chan packet.Packet) {
	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := tun.Read(buf)
		if err != nil {
			log.Error("Reading from tunnel:", err)
			return
		}

		log.Debug("received client packet:", buf[:n])
		p := packet.NewPacket()
		p.SetData(buf[:n])
		pOut <- *p
	}
}

/* handle traffic from target to client. */
func handleTunIn(tun *water.Interface, pIn chan packet.Packet) {
	for {
		p := <-pIn
		tun.Write(p.Data)
	}
}
