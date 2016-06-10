package ippool

import (
	"errors"
	"net"
)

var ErrNoIPAddrAvaliable = errors.New("No IPAddr avaliable")
var ErrIPAddrPoolFull = errors.New("IPAddrPool is full")

type IPAddrPool chan *net.IPNet

func NewIPAddrPool(reserveIP net.IP, ipNet *net.IPNet) (p IPAddrPool) {
	// calculate pool size
	ones, bits := ipNet.Mask.Size()
	size := 1<<uint(bits-ones) - 2

	// Initialize pool
	p = make(chan *net.IPNet, size)
	for identity := 1; identity <= size; identity++ {
		ip := make(net.IP, 4)
		copy(ip, ipNet.IP.To4())
		for i, index := 1, identity; index != 0; i++ {
			ip[len(ip)-i] = ip[len(ip)-i] | (byte)(0xFF&index)
			index = index >> 8
		}

		if ip.Equal(reserveIP) {
			continue
		}

		p <- &net.IPNet{ip, ipNet.Mask}
	}
	return
}

func (p IPAddrPool) Get() (ip *net.IPNet, err error) {
	select {
	case ip = <-p:
		return
	default:
		err = ErrNoIPAddrAvaliable
		return
	}
}

func (p IPAddrPool) Put(ip *net.IPNet) (err error) {
	select {
	case p <- ip:
		return
	default:
		err = ErrIPAddrPoolFull
		return
	}
}
