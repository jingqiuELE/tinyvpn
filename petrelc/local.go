package main

import (
	"fmt"
	"github.com/songgao/water"
	"packet"
	"tunnel"
)

/* Handle client's traffic, wrap each packet with outer IP Header */
func startListenTun(pIn, pOut chan packet.Packet) error {
	tun, err := water.NewTUN("")
	if err != nil {
		fmt.Printf("Error is %v\n", err)
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

func handleTunOut(tun *water.Interface, pOut chan packet.Packet) {
	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := tun.Read(buf)
		if err != nil {
			fmt.Println("Error reading from tunnel:", err)
			return
		}

		p := packet.NewPacket(buf[:n])
		pOut <- *p
	}
}

func handleTunIn(tun *water.Interface, pIn chan packet.Packet) {
	for {
		p := <-pIn
		buf, err := packet.Marshal(&p)
		if err != nil {
			fmt.Println("Failed to marshal packet:", err)
			continue
		}
		tun.Write(buf)
	}
}
