package main

import (
	"net"
	"packet"
	"session"
	"strconv"
	"sync"
)

type ConnServer struct {
	sync.RWMutex
	connMap map[session.Key]Connection
	eout    chan packet.Packet
	ein     chan packet.Packet
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

func (u UConnection) writePacket(p packet.Packet) error {
	buf := packet.MarshalToSlice(p)
	_, err := ProxyConn.WriteToUDP(buf, u.UDPAddr)
	if err != nil {
		log.Error("Failed to write Packet to UDP client:", err)
	}
	return err
}

func (t TConnection) writePacket(p packet.Packet) error {
	err := packet.MarshalToStream(p, t.TCPConn)
	if err != nil {
		log.Error("Failed to write Packet to TCP client:", err)
	}
	return err
}

func newConnServer(serverIP string, tcpPort int, udpPort int,
	eout chan packet.Packet, ein chan packet.Packet) (*ConnServer, error) {
	c := new(ConnServer)
	c.connMap = make(map[session.Key]Connection)
	c.eout = eout
	c.ein = ein

	if tcpPort != 0 {
		err := c.startTCPListener(serverIP, tcpPort)
		if err != nil {
			log.Error(err)
			return c, err
		}
	}

	if udpPort != 0 {
		err := c.startUDPListener(serverIP, udpPort)
		if err != nil {
			log.Error(err)
			return c, err
		}
	}

	go func() {
		for {
			p := <-c.eout
			sk, err := session.NewKey()
			if err != nil {
				continue
			}
			copy(sk[:], p.Header.Sk[:])

			conn, ok := c.connMap[*sk]
			if ok {
				conn.writePacket(p)
			}
		}
	}()
	return c, nil
}

func (c *ConnServer) startUDPListener(serverIP string, port int) error {
	serverAddr := serverIP + ":" + strconv.Itoa(port)
	listenAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		log.Error("Error when resoving UDP Address!")
		return err
	}

	pudp, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		log.Error(err)
		return err
	}

	ProxyConn = pudp

	go func() {
		for {
			p, err := c.readPacketFromUDP(pudp)
			if err != nil {
				return
			}
			c.ein <- p
		}
	}()
	return err
}

func (c *ConnServer) startTCPListener(serverIP string, port int) error {
	serverAddr := serverIP + ":" + strconv.Itoa(port)
	listenAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		log.Error(err)
		return err
	}

	ln, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		log.Error(err)
		return err
	}

	go func() {
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				log.Error(err)
				return
			}
			go c.handleTCPConn(conn)
		}
	}()
	return err
}

func (c *ConnServer) handleTCPConn(conn *net.TCPConn) error {
	p, err := readPacketFromTCP(conn)
	if err != nil {
		log.Error(err)
		return err
	}

	t := new(TConnection)
	t.TCPConn = conn

	sk := new(session.Key)
	copy(sk[:], p.Header.Sk[:])

	c.Lock()
	c.connMap[*sk] = t
	c.Unlock()

	c.ein <- p
	return err
}

func readPacketFromTCP(t *net.TCPConn) (packet.Packet, error) {
	p, err := packet.UnmarshalStream(t)
	if err != nil {
		log.Error("UnmarshalFromStream failed:", err)
	}
	return p, err
}

func (c *ConnServer) readPacketFromUDP(u *net.UDPConn) (packet.Packet, error) {
	var p packet.Packet
	buf := make([]byte, BUFFERSIZE)
	_, cliaddr, err := u.ReadFromUDP(buf)
	if err != nil {
		log.Error("reading from ", cliaddr.String(), err)
		return p, err
	}

	p, err = packet.UnmarshalSlice(buf)
	if err != nil {
		log.Error("creating Packet from buffer:", err)
		return p, err
	}

	uc := new(UConnection)
	uc.UDPAddr = cliaddr

	sk := new(session.Key)
	copy(sk[:], p.Header.Sk[:])

	c.Lock()
	c.connMap[*sk] = uc
	c.Unlock()

	return p, err
}
