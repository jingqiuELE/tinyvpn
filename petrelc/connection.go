package main

import (
	"github.com/jingqiuELE/tinyvpn/internal/packet"
	"net"
	"strconv"
	"strings"
)

func startConnection(serverAddr string, protocol string, port int, eOut, eIn chan packet.Packet) error {
	connServer := serverAddr + ":" + strconv.Itoa(port)

	if strings.Compare(protocol, "udp") == 0 {
		return startUDPConnection(connServer, port, eOut, eIn)
	} else {
		return startTCPConnection(connServer, port, eOut, eIn)
	}
}

func startUDPConnection(connServer string, port int, eOut, eIn chan packet.Packet) error {
	raddr, err := net.ResolveUDPAddr("udp", connServer)
	if err != nil {
		log.Error("Resolving connServer:", err)
		return err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Error("Dial remote:", err)
		return err
	}

	go handleUDPOut(conn, eOut)
	go handleUDPIn(conn, eIn)
	return nil
}

func startTCPConnection(connServer string, port int, eOut, eIn chan packet.Packet) error {
	raddr, err := net.ResolveTCPAddr("tcp", connServer)
	if err != nil {
		log.Error("Resolving connServer:", err)
		return err
	}

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		log.Error("Dial remote:", err)
		return err
	}

	go handleTCPOut(conn, eOut)
	go handleTCPIn(conn, eIn)
	return nil
}

/* traffic from client to target */
func handleUDPOut(conn *net.UDPConn, eOut chan packet.Packet) {
	for {
		p := <-eOut
		buf, err := packet.MarshalToSlice(p)
		if err != nil {
			log.Error("Failed to marshal packet:", err)
			continue
		}

		_, err = conn.Write(buf)
		if err != nil {
			log.Error("Writing to Connection:", err)
		}
	}
}

/* traffic from target to client */
func handleUDPIn(conn *net.UDPConn, eIn chan packet.Packet) {
	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Error("Read from Connection:", err)
			continue
		}

		p, err := packet.UnmarshalFromSlice(buf[:n])
		if err != nil {
			log.Error("Failed to unmarshal data:", err)
			continue
		}
		eIn <- p
	}
}

/* traffic from client to target */
func handleTCPOut(conn *net.TCPConn, eOut chan packet.Packet) {
	for {
		p := <-eOut
		err := packet.MarshalToStream(p, conn)
		if err != nil {
			log.Error("Failed to marshal packet:", err)
			continue
		}
	}
}

/* traffic from target to client */
func handleTCPIn(conn *net.TCPConn, eIn chan packet.Packet) {
	for {
		p, err := packet.UnmarshalFromStream(conn)
		if err != nil {
			log.Error("Failed to unmarshal data:", err)
			continue
		}
		eIn <- p
	}
}
