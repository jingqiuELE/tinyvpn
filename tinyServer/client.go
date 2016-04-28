package main

import (
    "net"
)

type Client struct{
    udpConn *net.UDPConn
    tcpConn *net.TCPConn
}

func (c *Client) Write(buf []byte) (n int, err error) {
    if c.udpConn != nil {
       n, err = c.udpConn.Write(buf)
    } else if c.tcpConn != nil {
       n, err = c.tcpConn.Write(buf)
    }
    return
}
