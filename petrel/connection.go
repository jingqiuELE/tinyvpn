package main

import (
	"encoding/binary"
	"fmt"
	"github.com/songgao/water/waterutil"
	"net"
	"strconv"
	"strings"
)

func StartUDPListener(serverIP string, port int, c chan Packet,
	sessionMap map[SessionKey]Session) error {
	serverAddr := serverIP + ":" + strconv.Itoa(port)
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

	var p Parket
	for {
		p, err := readPacketFromUDP(conn)
		if err != nil {
			fmt.Println("Error:reading from ", err, addr)
		} else {
			updateIpSessionMap(addr, pHeader.sessionKey)
			c <- p
		}
	}
	return err
}

func StartTCPListener(serverIP string, port int, c chan Packet,
	sessionMap map[SessionKey]Session) error {
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
			go handleTCPConn(conn, c)
		}
	}
	return err
}

func handleTCPConn(conn *net.TCPConn, c chan []byte) {
	p, err := readPacketFromTCP(conn)
	if err != nil {
		fmt.Println("Error:reading ", err)
	} else {
		updateIpSessionMap(addr, p.sessionKey)
		c <- p
	}
}

const packetHeaderLen = 16

/* readPacket would assume the buf always starts with beginning of a Packet. */
func readPacketFromTCP(conn *net.TCPConn, buf []byte) (p Packet, err error) {
	var p Packet
	reader := bufio.NewReader(conn)
	err := binary.Read(reader, binary.BigEndian, &p.Header)
	if err != nil {
		fmt.Println("binary read Packet Header failed:", err)
	}
	p.data = make([]byte, p.Header.length)
	err := binary.Read(data, binary.BigEndian, &p.Data)
	if err != nil {
		fmt.Println("binary read Packet Data failed:", err)
	}
	return p, err
}

func readPacketFromUDP(conn *net.UDPConn) (p Packet, addr *net.UDPAddr, err error) {
	var p Packet
	buf := make([]byte, BUFFERSIZE)
	len, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println("Error:reading from ", err, addr)
	} else {
		data := bytes.NewBuffer(buf[0:len])
		binary.Read(data, binary.BigEndian, &p.Header)
		p.data = make([]byte, p.Header.length)
		binary.Read(data, binary.BigEndian, &p.Data)
	}
	return p, addr, err
}
