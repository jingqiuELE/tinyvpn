package main

import (
	"bytes"
	"net"
	"strconv"
	"strings"
	"sync"
)

type ConnServer struct {
	sync.RWMutex
	connMap map[Index]Connection
	eOut    chan *Packet
	eIn     chan *Packet
}

type UConnection struct {
	UDPAddr *net.UDPAddr
}

type TConnection struct {
	TCPConn *net.TCPConn
}

type Connection interface {
	writePacket(p *Packet) error
}

var ProxyConn *net.UDPConn

func (u UConnection) writePacket(p *Packet) error {
	var buf bytes.Buffer
	err := Encode(p, &buf)
	if err != nil {
		return err
	}

	_, err = ProxyConn.WriteToUDP(buf.Bytes(), u.UDPAddr)
	if err != nil {
		log.Error("Failed to write Packet to UDP client:", err)
	}
	return err
}

func (t TConnection) writePacket(p *Packet) error {
	err := Encode(p, t.TCPConn)
	if err != nil {
		log.Error("Failed to write Packet to TCP client:", err)
	} else {
		log.Debug("packet send to client:", p)
	}

	return err
}

func newConnServer(serverIP string, tcpPort int, udpPort int,
	eOut, eIn chan *Packet) (*ConnServer, error) {
	c := &ConnServer{
		connMap: make(map[Index]Connection),
		eOut:    eOut,
		eIn:     eIn,
	}

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
		var sk Index
		for {
			p := <-c.eOut
			sk = p.Sk // Simple assignment would copy an array in go

			c.RLock()
			conn, ok := c.connMap[sk]
			c.RUnlock()
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
			c.eIn <- p
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
			log.Info("New tcp connection accepted.")
			go c.handleTCPConn(conn)
		}
	}()
	return err
}

func (c *ConnServer) handleTCPConn(conn *net.TCPConn) error {
	var err error
	for {
		p, err := readPacketFromTCP(conn)
		if err != nil {
			log.Error(err)
			break
		}

		t := new(TConnection)
		t.TCPConn = conn

		sk := p.Sk

		c.Lock()
		c.connMap[sk] = t
		c.Unlock()

		c.eIn <- p
	}
	return err
}

func readPacketFromTCP(t *net.TCPConn) (*Packet, error) {
	p, err := Decode(t)
	if err != nil {
		log.Error("UnmarshalFromStream failed:", err)
	}
	return p, err
}

func (c *ConnServer) readPacketFromUDP(u *net.UDPConn) (*Packet, error) {
	buf := make([]byte, PacketSize)
	_, cliaddr, err := u.ReadFromUDP(buf)
	if err != nil {
		log.Error("reading from ", cliaddr.String(), err)
		return nil, err
	}

	p, err := Decode(bytes.NewReader(buf))
	if err != nil {
		log.Error("creating Packet from buffer:", err)
		return p, err
	}

	uc := new(UConnection)
	uc.UDPAddr = cliaddr

	sk := p.Sk
	c.Lock()
	c.connMap[sk] = uc
	c.Unlock()

	return p, err
}

func startConnection(serverAddr string, protocol string, port int, eOut, eIn chan *Packet) error {
	connServer := serverAddr + ":" + strconv.Itoa(port)

	if strings.Compare(protocol, "udp") == 0 {
		return startUDPConnection(connServer, port, eOut, eIn)
	} else {
		return startTCPConnection(connServer, port, eOut, eIn)
	}
}

func startUDPConnection(connServer string, port int, eOut, eIn chan *Packet) error {
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

func startTCPConnection(connServer string, port int, eOut, eIn chan *Packet) error {
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
func handleUDPOut(conn *net.UDPConn, eOut chan *Packet) {
	for {
		p := <-eOut
		buf := new(bytes.Buffer)
		err := Encode(p, buf)
		if err != nil {
			log.Error("Failed to marshal packet:", err)
			continue
		}

		_, err = conn.Write(buf.Bytes())
		if err != nil {
			log.Error("Writing to Connection:", err)
		}
	}
}

/* traffic from target to client */
func handleUDPIn(conn *net.UDPConn, eIn chan *Packet) {
	for {
		// FIXME: Can we use UDPConn as a Reader directly?
		p, err := Decode(conn)
		if err != nil {
			log.Error("Read from Connection:", err)
			continue
		}
		eIn <- p
	}
}

/* traffic from client to target */
func handleTCPOut(conn *net.TCPConn, eOut chan *Packet) {
	for {
		p := <-eOut
		err := Encode(p, conn)
		if err != nil {
			log.Error("Failed to marshal packet:", err)
			continue
		}
	}
}

/* traffic from target to client */
func handleTCPIn(conn *net.TCPConn, eIn chan *Packet) {
	for {
		p, err := Decode(conn)
		if err != nil {
			log.Error("Failed to unmarshal data:", err)
			continue
		}
		eIn <- p
	}
}
