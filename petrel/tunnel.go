package main

import (
	"os/exec"
	"strconv"

	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
)

type Tunnel struct {
	Addr string
	Mtu  int
	tun  *water.Interface
	book Book
}

func startTUN(ip string, mtu int, book Book) (chan<- *Packet, <-chan *Packet, error) {
	tun := Tunnel{Addr: ip, Mtu: MTU, book: book}
	return tun.Start()
}

func (tunnel *Tunnel) Start() (chan<- *Packet, <-chan *Packet, error) {

	t, err := water.NewTUN("")
	if err != nil {
		return nil, nil, err
	}

	// Add IP address to tun interface
	_, err = exec.Command("ip", "addr", "add", tunnel.Addr, "dev", t.Name()).Output()
	if err != nil {
		return nil, nil, err
	}

	// Set its MTU
	_, err = exec.Command("ip", "link", "set", "dev", t.Name(), "mtu", strconv.Itoa(tunnel.Mtu)).Output()
	if err != nil {
		return nil, nil, err
	}

	// Bring up the device
	_, err = exec.Command("ip", "link", "set", "dev", t.Name(), "up").Output()
	if err != nil {
		return nil, nil, err
	}

	tunnel.tun = t
	in := tunnel.writeHandler(t)
	out := tunnel.readHandler(t)
	return in, out, err
}

func (tunnel *Tunnel) writeHandler(tun *water.Interface) chan<- *Packet {
	in := make(chan *Packet)
	go func() {
		for {
			p, ok := <-in
			if !ok {
				log.Error("Failed to read from pIn:")
				return
			}

			src_ip := waterutil.IPv4Source(p.Data)
			tunnel.book.Add(src_ip.String(), p.Sk)
			log.Debug("TUN SENDING: ", src_ip, p)

			_, err := tunnel.tun.Write(p.Data)
			if err != nil {
				log.Error("Error writing to tun!", err)
			}
		}
	}()
	return in
}

func (tunnel *Tunnel) readHandler(tun *water.Interface) <-chan *Packet {
	out := make(chan *Packet)
	go func() {
		for {
			buffer := make([]byte, MTU)
			n, err := tunnel.tun.Read(buffer)
			log.Debug("TUN READ:", buffer[:n])
			if err != nil {
				log.Error("Error reading from tunnel.")
				return
			}

			dst_ip := waterutil.IPv4Destination(buffer[:n])
			sk, ok := tunnel.book.getSession(dst_ip.String())
			// Missing session
			if !ok {
				log.Debug("ignoring packet: no session key found for ip:", dst_ip)
				log.Debug(buffer[:n])
				continue
			}

			p := new(Packet)
			p.Sk = sk
			p.Data = buffer[:n]
			log.Debug("TUN RECEIVED: ", p)
			out <- p
		}
	}()
	return out
}
