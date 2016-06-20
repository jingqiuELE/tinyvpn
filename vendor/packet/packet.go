package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"session"
)

const IvLen = 8

type PacketHeader struct {
	Iv  []byte //8 bytes slice
	Sk  []byte //6 bytes slice
	Len uint16 //2 bytes
}

type Packet struct {
	Header PacketHeader
	Data   []byte
}

func NewPacket() *Packet {
	p := new(Packet)
	p.Header.Iv = make([]byte, IvLen)
	p.Header.Sk = make([]byte, session.KeyLen)
	return p
}

func (p *Packet) SetData(data []byte) {
	p.Header.Len = uint16(len(data))
	p.Data = data
}

func UnmarshalSlice(buf []byte) (Packet, error) {
	var err error
	p := NewPacket()
	p.Header.Iv = buf[:8]
	p.Header.Sk = buf[8:14]

	dlen := buf[14:16]
	p.Header.Len = uint16(dlen[0])*256 + uint16(dlen[1])

	p.Data = buf[16:]
	if len(p.Data) != int(p.Header.Len) {
		fmt.Println("p.Len not equal to its data len!", len(p.Data), p.Header.Len)
		err = errors.New("p.Len not equal to its data len!")
	}
	return *p, err
}

func MarshalToSlice(p Packet) []byte {
	buf := p.Header.Iv
	buf = append(buf, p.Header.Sk...)

	dlen := make([]byte, 2)
	dlen[0] = byte(p.Header.Len / 256)
	dlen[1] = byte(p.Header.Len % 256)

	buf = append(buf, dlen...)
	buf = append(buf, p.Data...)
	return buf
}

func UnmarshalStream(s io.Reader) (Packet, error) {
	p := NewPacket()

	err := binary.Read(s, binary.BigEndian, &p.Header)
	if err != nil {
		fmt.Println("binary read Packet Header failed:", err)
	}

	p.Data = make([]byte, p.Header.Len)
	err = binary.Read(s, binary.BigEndian, &p.Data)
	if err != nil {
		fmt.Println("binary read Packet Data failed:", err)
	}

	return *p, err
}

func MarshalToStream(p Packet, writer io.Writer) error {
	err := binary.Write(writer, binary.BigEndian, p)
	if err != nil {
		fmt.Println("binary write Packet to stream failed:", err)
	}
	return err
}
