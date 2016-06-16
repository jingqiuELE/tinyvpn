package tunnel

import (
	"fmt"
	"github.com/codeskyblue/go-sh"
	"github.com/songgao/water"
)

func AddAddr(t *water.Interface, addr string) error {
	err := sh.Command("ip", "addr", "add", addr, "dev", t.Name()).Run()
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
