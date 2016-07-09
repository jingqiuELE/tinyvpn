package main

import (
	"encrypt"
	"packet"
	"session"
)

type EncryptServer struct {
	auth *AuthServer
	eOut chan packet.Packet
	eIn  chan packet.Packet
	pOut chan packet.Packet
	pIn  chan packet.Packet
}

func newEncryptServer(a *AuthServer, eOut, eIn, pOut, pIn chan packet.Packet) (*EncryptServer, error) {
	e := &EncryptServer{
		auth: a,
		eOut: eOut,
		eIn:  eIn,
		pOut: pOut,
		pIn:  pIn,
	}

	go e.start()
	return e, nil
}

/*
Direction:
In  -> from client to target
Out -> from target to client
*/
func (e *EncryptServer) start() {
	var p packet.Packet
	var secret session.Secret
	var ok bool
	var err error

	eIn_ok := true
	pOut_ok := true

	for {
		if !eIn_ok || !pOut_ok {
			log.Notice("channel closed!")
			return
		}
		select {
		case p, eIn_ok = <-e.eIn:
			secret, ok = e.auth.getSecret(p.Header.Sk)
			if ok != true {
				log.Error("eIn:Unknown session Index! skipping packet...")
				continue
			}
			err = encrypt.DecryptPacket(&p, secret[:])
			if err != nil {
				log.Error(err)
			}
			e.pIn <- p
		case p, pOut_ok = <-e.pOut:
			iv, _ := packet.NewIv()
			p.Header.Iv = *iv
			secret, ok = e.auth.getSecret(p.Header.Sk)
			if ok != true {
				log.Error("pOut:Unknown session Index! skipping packet...")
				continue
			}
			err = encrypt.EncryptPacket(&p, secret[:])
			if err != nil {
				log.Error(err)
			}
			e.eOut <- p
		}
	}
}
