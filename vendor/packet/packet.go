package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// After marshal, the length of packet header should be 8+6+2 bytes.
const PacketHeaderLen = 16

type PacketHeader struct {
	Iv     [8]byte
	Sk     [6]byte
	Length uint16
}

type Packet struct {
	Header PacketHeader
	Data   []byte
}

func NewPacket(data []byte) (p *Packet) {
	p = new(Packet)
	p.Header.Length = uint16(len(data))
	p.Data = data
	return
}

func Unmarshal(buf []byte) (p Packet, err error) {
	data := bytes.NewBuffer(buf[:])
	err = binary.Read(data, binary.BigEndian, &p.Header)
	if err != nil {
		return
	}
	p.Data = make([]byte, p.Header.Length)
	binary.Read(data, binary.BigEndian, &p.Data)
	return
}

func Marshal(p *Packet) (buf []byte, err error) {
	data := new(bytes.Buffer)
	err = binary.Write(data, binary.BigEndian, p.Header)
	if err != nil {
		fmt.Println("Failed to marshal packet header:", err)
		return
	}

	err = binary.Write(data, binary.BigEndian, p.Data)
	if err != nil {
		fmt.Println("Failed to marshal packet data:", err)
	}

	buf = data.Bytes()

	return
}
