package main

import (
	"packet"
	"session"
)

func startEncrypt(eOut, eIn, pOut, pIn chan packet.Packet, sk session.Key) error {
	var p packet.Packet
	eIn_ok := true
	pOut_ok := true
	go func() {
		for {
			if !eIn_ok || !pOut_ok {
				log.Notice("channel closed!")
				return
			}
			select {
			case p, eIn_ok = <-eIn:
				pIn <- p
			case p, pOut_ok = <-pOut:
				eOut <- p
			}
		}
	}()
	return nil
}
