package tunnel

import (
	"fmt"
	"github.com/codeskyblue/go-sh"
	"github.com/songgao/water"
	"net"
)

func AddAddr(t *water.Interface, vpnnet string) error {
	_, ipNet, err := net.ParseCIDR(vpnnet)
	if err != nil {
		fmt.Println("Error in vpnnet format: %V", vpnnet)
		return err
	}

	err = sh.Command("ip", "addr", "add", ipNet.String(), "dev", t.Name()).Run()
	if err != nil {
		fmt.Println("Error adding address to:", t.Name())
	}
	return err
}

func Bringup(t *water.Interface) error {
	err := sh.Command("ip", "link", "set", "dev", t.Name(), "up").Run()
	if err != nil {
		fmt.Println("Error seting up dev:", t.Name())
		return err
	}
	return err
}
