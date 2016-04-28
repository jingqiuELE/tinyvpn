package  main

import (
    "net"
    "fmt"
    "github.com/songgao/water"
    "github.com/codeskyblue/go-sh"
)

const BUFFERSIZE int = 1522

type Tunnel struct {
    vpnGateway net.IP
    vpnNet *net.IPNet
    ifce  *water.Interface
}

func CreateTunnel(vpnnet string) (*Tunnel, error) {
    ip, ipNet, err := net.ParseCIDR(vpnnet)
    if err != nil {
        fmt.Println("Error in vpnnet format: %V", vpnnet)
        return nil, nil
    }

    ifce, err := water.NewTUN("")
    if err != nil{
        fmt.Println("Error creating tun interface", err)
        return nil, err
    }
    tunnel := Tunnel{
                vpnGateway: ip,
                vpnNet: ipNet,
                ifce: ifce}
    fmt.Println("Created tun interface ", ifce)

    err = tunnel.Bringup()
    err = tunnel.AddAddr()
    SetNatRule()
    return &tunnel, err
}

func (t *Tunnel) Run(c chan []byte) {
    buffer := make([]byte, BUFFERSIZE)
    for {
        _, err := t.ifce.Read(buffer)
        if err != nil {
            fmt.Println("Error reading from tunnel.")
        }
        c <- buffer
    }
}

func (t *Tunnel) Bringup() error {
    return sh.Command("ip", "link", "set", "dev", t.ifce.Name(), "up").Run()
}

func (t *Tunnel) AddAddr() error {
    return sh.Command("ip", "addr", "add", t.vpnNet, "dev", t.ifce.Name()).Run()
}

func (t *Tunnel) Write(p []byte) (n int, err error) {
    n, err = t.ifce.Write(p)
    return n, err
}

func GetDefaultRouteIf() ([]byte, error) {
    return sh.Command("ip", "route", "show", "default").Command("awk", "'/default/ {print $5}'").Output()
}

func SetNatRule() error {
    var err error
    default_if, err := GetDefaultRouteIf()
    if err != nil {
        fmt.Println("Cannot get the default route interface")
    } else {
        err = sh.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-o", default_if, "-j", "MASQUERADE").Run()
    }
    return err
}
