package main

import (
	"packet"
	"session"
)

/*
Direction:
Out: from client to target
In: from target to client
*/
func startEncrypt(eOut, eIn, pOut, pIn chan packet.Packet, sk session.Key, secret session.Secret) (err error) {
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
				p.Header.Iv, err = packet.NewIv()
				if err != nil {
					log.Error(err)
				}
				p.Header.Sk = sk
				eOut <- p
			}
		}
	}()
	return
}
