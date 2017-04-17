package ippool

import (
	"fmt"
	"net"
	"testing"
)

func Test_Ippool(t *testing.T) {
	var in *net.IPNet
	var err error

	internalIP, ipNet, err := net.ParseCIDR("192.168.1.1/24")
	if err != nil {
		t.Error(err)
	}

	ones, bits := ipNet.Mask.Size()
	//don't count in all zero, all one, and the gateway addr.
	num_ip := 1<<uint(bits-ones) - 3
	ips := make([]*net.IPNet, num_ip)
	p := NewIPAddrPool(internalIP, ipNet)

	for index := range ips {
		in, err = p.Get()
		if in != nil {
			ips[index] = in
			fmt.Println(ips[index].IP.String())
		} else {
			t.Error(err)
			break
		}
	}

	for index := range ips {
		in = ips[index]
		if in != nil {
			err = p.Put(in)
			if err != nil {
				t.Error(err)
			}
		}
	}
}
