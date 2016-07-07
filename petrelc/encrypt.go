package main

import (
	"encrypt"
	"packet"
	"session"
)

/*
Direction:
Out: from client to target
In: from target to client
*/
func startEncrypt(eOut, eIn, pOut, pIn chan packet.Packet, sk session.Index, secret session.Secret) (err error) {
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
				err = encrypt.DecryptPacket(&p, secret[:])
				if err != nil {
					log.Error(err)
				}
				pIn <- p
			case p, pOut_ok = <-pOut:
				iv, err := packet.NewIv()
				if err != nil {
					log.Error("Failed to generate iv.")
					continue
				}
				p.Header.Iv = *iv
				p.Header.Sk = sk
				log.Debug("iv:", p.Header.Iv[:])
				log.Debug("sk:", p.Header.Sk[:])
				err = encrypt.EncryptPacket(&p, secret[:])
				if err != nil {
					log.Error(err)
				}
				eOut <- p
			}
		}
	}()
	return
}
