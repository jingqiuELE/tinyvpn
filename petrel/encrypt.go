package main

import (
	"fmt"
	"packet"
)

type EncryptServer struct {
	eOut chan packet.Packet
	eIn  chan packet.Packet
	pOut chan packet.Packet
	pIn  chan packet.Packet
}

func newEncryptServer(eOut, eIn, pOut, pIn chan packet.Packet) (*EncryptServer, error) {
	e := new(EncryptServer)
	e.eOut = eOut
	e.eIn = eIn
	e.pOut = pOut
	e.pIn = pIn

	go e.start()
	return e, nil
}

/*
Direction:
In  -> from client to vpn
Out -> from vpn to client
*/
func (e *EncryptServer) start() {
	var eIn_ok, pOut_ok bool
	var p packet.Packet
	for {
		if !eIn_ok || !pOut_ok {
			fmt.Println("channel closed!")
			return
		}
		select {
		case p, eIn_ok = <-e.eIn:
			e.pIn <- p
		case p, pOut_ok = <-e.pOut:
			e.eOut <- p
		}
	}
}
