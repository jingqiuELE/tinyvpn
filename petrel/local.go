package main

//// Handle client's traffic, wrap each packet with outer IP Header
//func startListenTun(pIn, pOut chan *Packet, ip net.IP) error {
//tun := Tunnel{ip.String(), MTU, nil}

//if err != nil {
//log.Error(err)
//return err
//}

//go handleTunOut(tun, pOut)
//go handleTunIn(tun, pIn)

//return nil
//}

//// handle traffic from client to target
//func handleTunOut(tun *water.Interface, pOut chan *Packet) {
//for {
//pr := PacketReader{tun}
//p, err := pr.NextPacket()
//if err != nil {
//log.Error("Reading from tunnel:", err)
//return
//}

//pOut <- p
//}
//}

//// handle traffic from target to client.
//func handleTunIn(tun *water.Interface, pIn chan *Packet) {
//for {
//p := <-pIn
//tun.Write(p.Data)
//}
//}
