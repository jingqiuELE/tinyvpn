package main

import (
    "net"
    "fmt"
)

type TinyServer struct {
    secret string
    tunnel *Tunnel
    carrierProtocol string
    serverAddr string
}


//create vpn server, which in turn would create the tunnel interface and bring up it with NAT
//settings.
func CreateTinyServer(secret string, carrierProtocol string, serverAddr string, vpnnet string) (*TinyServer, error) {
    var err error
    var server TinyServer

    server.secret = secret
    server.carrierProtocol = carrierProtocol
    server.serverAddr = serverAddr
    server.tunnel, err = CreateTunnel(vpnnet)
    return &server, err
}

//vpn server starts handling packets.
func (s *TinyServer) Run() error {
    var err error

    ct := make(chan[]byte)
    go s.tunnel.Run(ct)

    cw := make(chan[]byte)
    switch s.carrierProtocol {
        case "udp":
            go StartUDPServer(s.serverAddr, cw)
            fmt.Println("Started UDP service.")
        case "tcp":
            go StartTCPServer(s.serverAddr, cw)
            fmt.Println("Started TCP service.")
        default:
            fmt.Println("tinyvpn can only run on udp or tcp protocol.")
    }

    for {
        select {
            case tunnelData := <-ct:
                 err = s.handleTunnelConn(tunnelData)
            case wanData := <-cw:
                 err = s.handleWanData(wanData)
        }
    }
    return err
}

//Destroy vpn server instance.
func (s *TinyServer) Close() error {
    return nil
}

func StartUDPServer(serverAddr string, c chan[]byte) error {
     listenAddr, err := net.ResolveUDPAddr("udp", serverAddr)
     if err != nil {
         fmt.Println("Error when resoving UDP Address!")
         return err
     }

     conn, err := net.ListenUDP("udp", listenAddr)
     if err != nil {
         fmt.Println("Error when listening to UDP Address!")
         return err
     }
     defer conn.Close()

     buf := make([]byte, BUFFERSIZE)
     for {
         n, addr, err := conn.ReadFromUDP(buf)
         if err != nil {
             fmt.Println("Error:reading from ", err, addr)
         } else {
             //Record connection to connRecord
             c <- buf
         }
     }

     return err
}

func StartTCPServer(serverAddr string, c chan[]byte) error {
     listenAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
     if err != nil {
         fmt.Println("Error when resoving TCP Address!")
         return err
     }

     ln, err := net.ListenTCP("tcp", listenAddr)
     if err != nil {
         fmt.Println("Error when listening to TCP Address!")
         return err
     }
     defer ln.Close()

     buf := make([]byte, BUFFERSIZE)
     for {
         conn, err := ln.Accept()
         if err != nil {
             fmt.Println("Error: ", err)
         } else {
             go handleTCPConn(conn, c)
         }
    }
    return err
}

func handleTCPConn(conn *net.Conn, c chan[]byte) {
    //Record connection to connRecord
    n, err := conn.Read(buf)
    if err != nil {
        fmt.Println("Error:reading from ", err, addr)
    } else {
        c <- buf
    }
}

func (s *TinyServer)handleWanData(buf []byte) {
   dst := waterUtil.IPv4Destination(buf)
   src := waterUtil.IPv4Source(buf)
   if vpnConn := connRecord[dst.String()]; vpnConn != nil {
       _, err := vpnConn.Write(buf)
   } else {
       //TODO: decrypt the packet
       s.tunnel.Write(buf)
   }
}

func (s *TinyServer)handleTunnelData(buf []byte) {
    dst := waterUtil.IPv4Destination(buf)
    if vpnConn := connRecord[dst.String()]; vpnConn != nil {
        _, err := vpnConn.Write(buf)
        if err != nil {
            fmt.Println("Failed to handle incomming tun data!")
        }
    }
}
