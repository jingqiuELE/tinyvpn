package packet

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/jingqiuELE/tinyvpn/internal/logger"
	"github.com/jingqiuELE/tinyvpn/internal/session"
	"github.com/op/go-logging"
)

var log = logger.Get(logging.ERROR)

//Assuming ethernet network interface has MTU 1500 bytes, from which we reduce
//the size of IPv4 header(minimal 20 bytes) and UDP header(8 bytes), to get the
//allowed packet size of petrel. The purpose of this is to fit a petrel packet in
//one frame without IP fragmentation.
const PacketSize = 1500 - 28

const IvLen = 12
const PacketHeaderSize = IvLen + session.IndexLen + 2
const MTU = PacketSize - PacketHeaderSize

type Iv [IvLen]byte

type Packet struct {
	Iv   Iv
	Sk   session.Index
	Len  uint16
	Data []byte
}

func (p Packet) String() string {
	return fmt.Sprintf("Iv: [% x], Sk: [% x], len: %d", p.Iv, p.Sk, p.Len, p.Data[:8])
}

func Decode(r io.Reader) (Packet, error) {
	var p Packet

	_, err := io.ReadFull(r, p.Iv[:])
	if err != nil {
		log.Error("failed to read initialization vector Iv:", err)
		return p, err
	}

	_, err = io.ReadFull(r, p.Sk[:])
	if err != nil {
		log.Error("failed to read session key Sk:", err)
		return p, err
	}

	err = binary.Read(r, binary.BigEndian, &p.Len)
	if err != nil {
		log.Error("binary read p.Header failed:", err)
		return p, err
	}

	p.Data = make([]byte, p.Len)
	n, err := io.ReadFull(r, p.Data)
	if err != nil {
		log.Error("read %d bytes", n)
		log.Error(err)
	}
	return p, err
}

func Encode(p Packet, w io.Writer) error {
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
