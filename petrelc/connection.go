package main

import (
	"net"
	"packet"
	"strconv"
)

func startConnection(serverAddr string, port int, eOut, eIn chan packet.Packet) error {
	connServer := serverAddr + ":" + strconv.Itoa(port)
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

	go handleOut(conn, eOut)
	go handleIn(conn, eIn)

	return err
}

/* traffic from client to target */
func handleOut(conn *net.UDPConn, eOut chan packet.Packet) {
	for {
		p := <-eOut
		log.Debug("client conn: iv:", p.Header.Iv[:])
		log.Debug("client conn: sk:", p.Header.Sk[:])
		buf, err := packet.MarshalToSlice(p)
		if err != nil {
			log.Error("Failed to marshal packet:", err)
			continue
		}

		log.Debug("client conn: buf:", buf)
		_, err = conn.Write(buf)
		if err != nil {
			log.Error("Writing to Connection:", err)
		}
	}
}

/* traffic from target to client */
func handleIn(conn *net.UDPConn, eIn chan packet.Packet) {
	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Error("Read from Connection:", err)
			continue
		}

		log.Debug("Connection handleIn:", buf[:n])

		p, err := packet.UnmarshalFromSlice(buf[:n])
		if err != nil {
			log.Error("Failed to unmarshal data:", err)
			continue
		}
		eIn <- p
	}
}
