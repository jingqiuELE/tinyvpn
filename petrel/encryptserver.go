package main

type EncryptServer struct {
	auth *AuthServer
	eOut chan *Packet
	eIn  chan *Packet
	pOut chan *Packet
	pIn  chan *Packet
}

func newEncryptServer(a *AuthServer, eOut, eIn, pOut, pIn chan *Packet) (*EncryptServer, error) {
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
	var secret Secret
	var ok bool
	var err error

	eIn_ok := true
	pOut_ok := true
	var p *Packet

	for {
		if !eIn_ok || !pOut_ok {
			log.Notice("channel closed!")
			return
		}
		select {
		case p, eIn_ok = <-e.eIn:
			secret, ok = e.auth.getSecret(p.Sk)
			if ok != true {
				log.Error("eIn:Unknown session Index! skipping ..")
				continue
			}
			err = DecryptPacket(p, secret[:])
			if err != nil {
				log.Error(err)
			}
			e.pIn <- p
		case p, pOut_ok = <-e.pOut:
			secret, ok = e.auth.getSecret(p.Sk)
			if ok != true {
				log.Error("pOut:Unknown session Index! skipping ..")
				continue
			}
			err = EncryptPacket(p, secret[:])
			if err != nil {
				log.Error(err)
			}
			e.eOut <- p
		}
	}
}

func startEncrypt(eOut, eIn, pOut, pIn chan *Packet, sk Index, secret Secret) (err error) {

	var p *Packet
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
				err = DecryptPacket(p, secret[:])
				if err != nil {
					log.Error(err)
				}
				pIn <- p
			case p, pOut_ok = <-pOut:
				p.Sk = sk
				err = EncryptPacket(p, secret[:])
				if err != nil {
					log.Error(err)
				}
				eOut <- p
			}
		}
	}()
	return
}
