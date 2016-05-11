package main

import (
	"bytes"
	"encoding/binary"
)

type PacketHeader struct {
	Iv         [8]byte
	sessionKey SessionKey
	Length     uint16
}

type Packet struct {
	Header PacketHeader
	Data   []byte
}

func NewPacket(buf []byte) (p Packet, err error) {
	data := bytes.NewBuffer(buf[:])
	err = binary.Read(data, binary.BigEndian, &p.Header)
	if err != nil {
		return
	}
	p.Data = make([]byte, p.Header.Length)
	binary.Read(data, binary.BigEndian, &p.Data)
	return
}
