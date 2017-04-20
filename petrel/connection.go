package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
)

type ConnServer struct {
	sync.RWMutex
	connMap map[Index]Connection
}

type UConnection struct {
	Conn *net.UDPConn
	Addr *net.UDPAddr
	Pr   *PacketReader
}

type TConnection struct {
	Conn *net.TCPConn
	Pr   *PacketReader
}

type Connection interface {
	writePacket(p *Packet) error
	readPacket() (*Packet, error)
}

func (u UConnection) writePacket(p *Packet) error {
	var buf bytes.Buffer
	err := p.Encode(&buf)
	if err != nil {
		return err
	}

	_, err = u.Conn.WriteToUDP(buf.Bytes(), u.Addr)
	return err
}

func (u UConnection) readPacket() (*Packet, error) {
	return u.Pr.NextPacket()
}

func (t TConnection) writePacket(p *Packet) error {
	return p.Encode(t.Conn)
}

func (t TConnection) readPacket() (*Packet, error) {
	return t.Pr.NextPacket()
}

func startConnServer(serverIP string, tcpPort, udpPort int) (chan<- *Packet, <-chan *Packet, error) {
	c := &ConnServer{connMap: make(map[Index]Connection)}
	fromClient := make(chan *Packet)
	var fromTcpClient, fromUdpClient <-chan *Packet
	var err error

	if tcpPort != 0 {
		fromTcpClient, err = c.startTCPListener(serverIP, tcpPort)
		if err != nil {
			log.Error(err)
			return nil, nil, err
		}
	}

	if udpPort != 0 {
		fromUdpClient, err = c.startUDPListener(serverIP, udpPort)
		if err != nil {
			log.Error(err)
			return nil, nil, err
		}
	}

	go func() {
		for {
			select {
			case p := <-fromTcpClient:
				fromClient <- p
			case p := <-fromUdpClient:
				fromClient <- p
			}
		}
	}()

	toClient := c.handleOut()
	return toClient, fromClient, nil
}

func (c *ConnServer) handleOut() chan<- *Packet {
	toClient := make(chan *Packet)
	go func() {
		for {
			p := <-toClient

			c.RLock()
			conn, ok := c.connMap[p.Sk]
			c.RUnlock()

			if ok {
				conn.writePacket(p)
			}
		}
	}()
	return toClient
}

func (c *ConnServer) startUDPListener(serverIP string, port int) (<-chan *Packet, error) {
	serverAddr := serverIP + ":" + strconv.Itoa(port)
	listenAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		log.Error("Error when resoving UDP Address!")
		return nil, err
	}

	pudp, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	fromClient := make(chan *Packet)
	go func() {
		for {
			p, err := c.readPacketFromUDP(pudp)
			if err != nil {
				return
			}
			fromClient <- p
		}
	}()
	return fromClient, err
}

func (c *ConnServer) startTCPListener(serverIP string, port int) (<-chan *Packet, error) {
	serverAddr := serverIP + ":" + strconv.Itoa(port)
	listenAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	ln, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	fromClient := make(chan *Packet)
	go func() {
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				log.Error(err)
				return
			}
			log.Info("New tcp connection accepted.")
			go c.handleTCPConn(conn, fromClient)
		}
	}()
	return fromClient, err
}

func (c *ConnServer) handleTCPConn(conn *net.TCPConn, fromClient chan<- *Packet) {
	t := TConnection{Conn: conn, Pr: &PacketReader{conn}}
	for {
		p, err := t.readPacket()
		if err != nil {
			log.Error(err)
			return
		}

		log.Debug("CONN RECEIVED:", p)
		c.Lock()
		c.connMap[p.Sk] = t
		c.Unlock()

		fromClient <- p
	}
}

func (c *ConnServer) readPacketFromUDP(u *net.UDPConn) (*Packet, error) {
	buf := make([]byte, PacketSize)
	_, cliaddr, err := u.ReadFromUDP(buf)
	if err != nil {
		log.Error("reading from ", cliaddr.String(), err)
		return nil, err
	}

	pr := PacketReader{bytes.NewReader(buf)}
	p, err := pr.NextPacket()
	if err != nil {
		log.Error("creating Packet from buffer:", err)
		return p, err
	}

	// FIXME: Does this mean a new UConnection is created for every UDP packet?
	uc := new(UConnection)
	uc.Addr = cliaddr

	sk := p.Sk
	c.Lock()
	c.connMap[sk] = uc
	c.Unlock()

	return p, err
}

func startConnection(serverAddr string, protocol string, port int) (chan<- *Packet, <-chan *Packet, error) {
	connServer := serverAddr + ":" + strconv.Itoa(port)

	var c Connection
	var err error
	if protocol == "udp" {
		c, err = createUDPConnection(connServer)
	} else if protocol == "tcp" {
		c, err = createTCPConnection(connServer)
	} else {
		return nil, nil, errors.New(fmt.Sprintf("unknown protocol %v", protocol))
	}

	if err != nil {
		return nil, nil, err
	}

	toServer := handleOut(c)
	fromServer := handleIn(c)
	return toServer, fromServer, nil
}

func createTCPConnection(connServer string) (Connection, error) {
	raddr, err := net.ResolveTCPAddr("tcp", connServer)
	if err != nil {
		log.Error("Resolving connServer:", err)
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		log.Error("Dial remote:", err)
		return nil, err
	}
	return TConnection{conn, &PacketReader{conn}}, nil
}

func createUDPConnection(connServer string) (Connection, error) {
	raddr, err := net.ResolveUDPAddr("udp", connServer)
	if err != nil {
		log.Error("Resolving connServer:", err)
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Error("Dial remote:", err)
		return nil, err
	}
	return UConnection{conn, raddr, &PacketReader{conn}}, nil
}

func handleOut(c Connection) chan<- *Packet {
	toRemote := make(chan *Packet, 100)
	log.Debug("CONN OUT START", toRemote, len(toRemote), cap(toRemote))
	go func() {
		for {
			p := <-toRemote
			log.Debug("CONN HOUT: got packet to send", p, toRemote, len(toRemote), cap(toRemote))
			err := c.writePacket(p)
			if err != nil {
				log.Error("Writing to Connection:", err)
			}
		}
	}()
	return toRemote
}

func handleIn(c Connection) <-chan *Packet {
	fromRemote := make(chan *Packet)
	log.Debug("CONN IN START", fromRemote)
	go func() {
		for {
			p, err := c.readPacket()
			log.Debug("CONN H_IN: received packet", p)
			if err != nil {
				log.Error("Read from Connection:", err)
				continue
			}
			fromRemote <- p
		}
	}()
	return fromRemote
}
