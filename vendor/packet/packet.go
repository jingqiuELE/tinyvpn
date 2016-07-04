package packet

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"session"
)

const IvLen = 8

/* Fixed size PacketHeader */
type PacketHeader struct {
	Iv  [IvLen]byte
	Sk  session.Index
	Len uint16
}

type Packet struct {
	Header PacketHeader
	Data   []byte
}

func NewPacket() *Packet {
	p := new(Packet)
	return p
}

func NewIv() ([IvLen]byte, error) {
	var iv [IvLen]byte
	_, err := rand.Read(iv[:])
	if err != nil {
		fmt.Println("Error:", err)
		return iv, err
	}
	return iv, err
}

func (p *Packet) SetData(data []byte) {
	p.Header.Len = uint16(len(data))
	p.Data = data
}

func UnmarshalFromSlice(buf []byte) (Packet, error) {
	reader := bytes.NewReader(buf)
	return UnmarshalFromStream(reader)
}

func MarshalToSlice(p Packet) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := MarshalToStream(p, buf)
	return buf.Bytes(), err
}

func UnmarshalFromStream(r io.Reader) (Packet, error) {
	p := NewPacket()

	err := binary.Read(r, binary.BigEndian, &p.Header)
	if err != nil {
		fmt.Println("binary read Packet Header failed:", err)
		return *p, err
	}

	p.Data = make([]byte, p.Header.Len)
	err = binary.Read(r, binary.BigEndian, &p.Data)
	if err != nil {
		fmt.Println("binary read Packet Data failed:", err)
	}

	return *p, err
}

func MarshalToStream(p Packet, w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, p.Header)
	if err != nil {
		fmt.Println("binary write Packet Header to stream failed:", err)
		return err
	}

	err = binary.Write(w, binary.BigEndian, p.Data)
	if err != nil {
		fmt.Println("binary read Packet Data failed:", err)
	}
	return err
}
