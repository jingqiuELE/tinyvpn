package main

import (
	"fmt"
	"packet"
	"session"
)

func startEncrypt(eOut, eIn, pOut, pIn chan packet.Packet, sk session.Index) error {
	var p packet.Packet
	go func() {
		var eIn_ok, pOut_ok bool
		for {
			if !eIn_ok || !pOut_ok {
				fmt.Println("channel closed!")
				return
			}
			select {
			case p, pOut_ok = <-pOut:
				eOut <- p
			case p, eIn_ok = <-eIn:
				pIn <- p
			}
		}
	}()
	return nil
}
