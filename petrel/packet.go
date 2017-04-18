package main

import (
	"encoding/binary"
	"fmt"
	"io"
)

//Assuming ethernet network interface has MTU 1500 bytes, from which we reduce
//the size of IPv4 header(minimal 20 bytes) and UDP header(8 bytes), to get the
//allowed packet size of petrel. The purpose of this is to fit a petrel packet in
//one frame without IP fragmentation.
const PacketSize = 1500 - 28

const IvLen = 12
const PacketHeaderSize = IvLen + IndexLen + 2
const MTU = PacketSize - PacketHeaderSize

type Iv [IvLen]byte

type Packet struct {
	Iv   Iv
	Sk   Index
	Len  uint16
	Data []byte
}

func (p Packet) String() string {
	return fmt.Sprintf("Iv: [% x], Sk: [% x], len: %d, Data[% x]", p.Iv, p.Sk, p.Len, p.Data)
}

func Decode(r io.Reader) (*Packet, error) {
	var p Packet
	_, err := io.ReadFull(r, p.Iv[:])
	if err != nil {
		log.Error("failed to read initialization vector Iv:", err)
		return nil, err
	}

	_, err = io.ReadFull(r, p.Sk[:])
	if err != nil {
		log.Error("failed to read session key Sk:", err)
		return nil, err
	}

	err = binary.Read(r, binary.BigEndian, &p.Len)
	if err != nil {
		log.Error("binary read p.Header failed:", err)
		return nil, err
	}

	p.Data = make([]byte, p.Len)
	n, err := io.ReadFull(r, p.Data)
	if err != nil {
		log.Error("read %d bytes", n)
		log.Error(err)
	}
	return &p, err
}

func Encode(p *Packet, w io.Writer) error {
	_, err := w.Write(p.Iv[:])
	if err != nil {
		log.Error("failed to write Iv:", err)
		return err
	}

	_, err = w.Write(p.Sk[:])
	if err != nil {
		log.Error("failed to write Sk:", err)
		return err
	}

	err = binary.Write(w, binary.BigEndian, p.Len)
	if err != nil {
		log.Error("binary write Packet data lengh to stream failed:", err)
		return err
	}

	_, err = w.Write(p.Data)
	if err != nil {
		log.Error("binary read Packet Data failed:", err)
	}
	return err
}
