package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/jingqiuELE/tinyvpn/internal/logger"
	"github.com/jingqiuELE/tinyvpn/internal/session"
	"github.com/op/go-logging"
	"io"
)

var log = logger.Get(logging.ERROR)

//Assuming ethernet network interface has MTU 1500 bytes, from which we reduce
//the size of IPv4 header(minimal 20 bytes) and UDP header(8 bytes), to get the
//allowed packet size of petrel. The purpose of this is to fit a petrel packet in
//one frame without IP fragmentation.
const PacketSize = 1500 - 28

const IvLen = 4
const PacketHeaderSize = IvLen + session.IndexLen + 2
const MTU = PacketSize - PacketHeaderSize

type Iv [IvLen]byte

/* Fixed-size PacketHeader to ease marshal and unmarshal */
type PacketHeader struct {
	Iv  Iv
	Sk  session.Index
	Len uint16
}

type Packet struct {
	Header PacketHeader
	Data   []byte
}

func (h PacketHeader) String() string {
	return fmt.Sprintf("Iv: [% x], Sk: [% x], len: %d", h.Iv, h.Sk, h.Len)
}

func NewPacket() *Packet {
	p := new(Packet)
	return p
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
		log.Error("Binary read p.Header failed:", err)
		return *p, err
	}
	log.Debug("UnmarshalFromStream-->p.Header:", p.Header)

	p.Data = make([]byte, p.Header.Len)
	n, err := io.ReadFull(r, p.Data)
	if err != nil {
		log.Error("Read %d bytes", n)
		log.Error(err)
	}
	return *p, err
}

func MarshalToStream(p Packet, w io.Writer) error {
	err := binary.Write(w, binary.BigEndian, p.Header)
	if err != nil {
		log.Error("binary write Packet Header to stream failed:", err)
		return err
	}

	err = binary.Write(w, binary.BigEndian, p.Data)
	if err != nil {
		log.Error("binary read Packet Data failed:", err)
	}
	return err
}
