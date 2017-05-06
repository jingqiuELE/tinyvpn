package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

//Assuming ethernet network interface has MTU 1500 bytes, from which we reduce
//the size of IPv4 header(minimal 20 bytes) and TCP header(32 bytes), to get the
//allowed packet size of petrel. The purpose of this is to fit a petrel packet in
//one frame without IP fragmentation.
const IPHeaderLen = 20
const TCPHeaderLen = 32
const UDPHeaderLen = 8

const IvLen = 12
const PacketHeaderSize = IvLen + IndexLen + 2
const GCMAuthTagLen = 16

const EthernetMTU = 1500
const MTU = EthernetMTU - IPHeaderLen - TCPHeaderLen - PacketHeaderSize - GCMAuthTagLen

const PacketSize = MTU + IPHeaderLen + UDPHeaderLen

type Iv []byte

type Packet struct {
	Iv            Iv
	Sk            Index
	Data          []byte
	EncryptedData []byte
}

type PacketReader struct {
	io.Reader
}

func (p Packet) String() string {
	t := len(p.Data)
	hasMore := ""
	if t > 8 {
		t = 8
		hasMore = "..."
	}
	return fmt.Sprintf("Iv: [% x], Sk: [% x], len: %d, Data[% x]%v", p.Iv, p.Sk, len(p.Data), p.Data[:t], hasMore)
}

func (r *PacketReader) NextPacket() (*Packet, error) {
	var p Packet
	p.Iv = make([]byte, IvLen)
	_, err := io.ReadFull(r, p.Iv)
	if err != nil {
		log.Error("failed to read initialization vector Iv:", err)
		return nil, err
	}

	_, err = io.ReadFull(r, p.Sk[:])
	if err != nil {
		log.Error("failed to read session key Sk:", err)
		return nil, err
	}

	var len uint16
	err = binary.Read(r, binary.BigEndian, &len)
	if err != nil {
		log.Error("binary read p.Header failed:", err)
		return nil, err
	}

	p.EncryptedData = make([]byte, len)
	n, err := io.ReadFull(r, p.EncryptedData)
	if err != nil {
		log.Error("read %d bytes", n)
		log.Error(err)
	}
	return &p, err
}

func (p *Packet) Encode(w io.Writer) (err error) {

	buf := new(bytes.Buffer)
	_, err = buf.Write(p.Iv[:])
	if err != nil {
		log.Error("failed to write Iv:", err)
		return err
	}

	_, err = buf.Write(p.Sk[:])
	if err != nil {
		log.Error("failed to write Sk:", err)
		return err
	}

	len := uint16(len(p.EncryptedData))
	err = binary.Write(buf, binary.BigEndian, len)
	if err != nil {
		log.Error("binary write Packet data lengh to stream failed:", err)
		return err
	}

	_, err = buf.Write(p.EncryptedData)
	if err != nil {
		log.Error("binary read Packet Data failed:", err)
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		log.Error("Failed to write buffer to interface", err)
	}

	return err
}
