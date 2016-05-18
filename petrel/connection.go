package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"packet"
	"strconv"
)

type ConnServer struct {
	connMap map[SessionKey]Connection
	out     chan packet.Packet
	in      chan packet.Packet
}

type UConnection struct {
	UDPAddr *net.UDPAddr
}

type TConnection struct {
	TCPConn *net.TCPConn
}

type Connection interface {
	writePacket(p packet.Packet) error
}

var ProxyConn *net.UDPConn

func (u UConnection) writePacket(p packet.Packet) (err error) {
	_, err = ProxyConn.WriteToUDP(p.Data, u.UDPAddr)
	return
}

func (t TConnection) writePacket(p packet.Packet) (err error) {
	conn := t.TCPConn
	_, err = conn.Write(p.Data)
	return
}

func newConnServer(serverIP string, tcpPort int, udpPort int,
	out chan packet.Packet, in chan packet.Packet) (*ConnServer, error) {
	c := new(ConnServer)
	c.connMap = make(map[SessionKey]Connection)
	c.out = out
	c.in = in
	if tcpPort != 0 {
		err := c.startTCPListener(serverIP, tcpPort)
		if err != nil {
			fmt.Printf("Error is %v\n", err)
			return c, err
		}
	}
	if udpPort != 0 {
		err := c.startUDPListener(serverIP, udpPort)
		if err != nil {
			fmt.Printf("Error is %v\n", err)
			return c, err
		}
	}
	return c, nil
}

func (c *ConnServer) startUDPListener(serverIP string, port int) error {
	serverAddr := serverIP + ":" + strconv.Itoa(port)
	listenAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		fmt.Println("Error when resoving UDP Address!")
		return err
	}

	pudp, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		fmt.Println("Error when listening to UDP Address!")
		return err
	}

	ProxyConn = pudp
	defer pudp.Close()

	for {
		p, err := c.readPacketFromUDP(pudp)
		if err != nil {
			return err
		}
		c.in <- p
	}
	return err
}

func (c *ConnServer) startTCPListener(serverIP string, port int) error {
	serverAddr := serverIP + ":" + strconv.Itoa(port)
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

	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			fmt.Println("Error: ", err)
			return err
		}
		go c.handleTCPConn(conn)
	}
	return err
}

func (c *ConnServer) handleTCPConn(conn *net.TCPConn) error {
	p, err := readPacketFromTCP(conn)
	if err != nil {
		fmt.Println("Error:reading ", err)
		return err
	}
	t := new(TConnection)
	t.TCPConn = conn
	c.connMap[p.Header.Sk] = t
	c.in <- p
	return err
}

const packetHeaderLen = 16

/* readPacket would assume the buf always starts with beginning of a Packet. */
func readPacketFromTCP(conn *net.TCPConn) (p packet.Packet, err error) {
	reader := bufio.NewReader(conn)
	err = binary.Read(reader, binary.BigEndian, &p.Header)
	if err != nil {
		fmt.Println("binary read Packet Header failed:", err)
	}
	p.Data = make([]byte, p.Header.Length)
	err = binary.Read(reader, binary.BigEndian, &p.Data)
	if err != nil {
		fmt.Println("binary read Packet Data failed:", err)
	}
	return p, err
}

const BUFFERSIZE = 1500

func (c *ConnServer) readPacketFromUDP(u *net.UDPConn) (p packet.Packet, err error) {
	buf := make([]byte, BUFFERSIZE)
	_, cliaddr, err := u.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("Error:reading from ", err, cliaddr.String())
		return
	}
	p, err = packet.Unmarshal(buf)
	if err != nil {
		fmt.Println("Error creating Packet from buffer!")
		return
	}

	uc := new(UConnection)
	uc.UDPAddr = cliaddr
	c.connMap[p.Header.Sk] = uc
	return p, err
}
