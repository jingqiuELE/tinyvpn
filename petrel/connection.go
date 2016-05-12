package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type UConnection net.UDPAddr
type TConnection net.TCPConn

type Connection interface {
	writePacket(p Packet) error
}

var ProxyConn *net.UDPConn

func (u UConnection) writePacket(p Packet) (err error) {
	addr := net.UDPAddr(u)
	_, err = ProxyConn.WriteToUDP(p.Data, &addr)
	return
}

func (t TConnection) writePacket(p Packet) (err error) {
	conn := net.TCPConn(t)
	_, err = conn.Write(p.Data)
	return
}

func startUDPListener(serverIP string, port int, c chan Packet,
	s *SessionMap) (err error) {
	serverAddr := serverIP + ":" + strconv.Itoa(port)
	listenAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		fmt.Println("Error when resoving UDP Address!")
		return
	}

	pudp, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		fmt.Println("Error when listening to UDP Address!")
		return
	}

	ProxyConn = pudp
	defer pudp.Close()

	for {
		p, err := readPacketFromUDP(pudp, s)
		if err != nil {
			return err
		}
		c <- p
	}
	return
}

func startTCPListener(serverIP string, port int, c chan Packet,
	s *SessionMap) error {
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
		} else {
			go handleTCPConn(conn, c, s)
		}
	}
	return err
}

func handleTCPConn(conn *net.TCPConn, c chan Packet, s *SessionMap) {
	p, err := readPacketFromTCP(conn)
	if err != nil {
		fmt.Println("Error:reading ", err)
	} else {
		var t TConnection = TConnection(*conn)
		s.Update(t, p.Header.sessionKey)
		c <- p
	}
}

const packetHeaderLen = 16

/* readPacket would assume the buf always starts with beginning of a Packet. */
func readPacketFromTCP(conn *net.TCPConn) (p Packet, err error) {
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

func readPacketFromUDP(u *net.UDPConn, s *SessionMap) (p Packet, err error) {
	buf := make([]byte, BUFFERSIZE)
	_, cliaddr, err := u.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("Error:reading from ", err, cliaddr.String())
		return
	}
	p, err = NewPacket(buf)
	if err != nil {
		fmt.Println("Error creating Packet from buffer!")
		return
	}

	c := UConnection(*cliaddr)
	s.Update(c, p.Header.sessionKey)
	return p, err
}
