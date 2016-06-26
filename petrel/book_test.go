package main

import (
	"bytes"
	"packet"
	"testing"

	"github.com/songgao/water"
)

type TestTunReadWriteCloser struct {
	lastWritten []byte
	returnChan  chan []byte
}

func (rwc TestTunReadWriteCloser) Read(p []byte) (n int, err error) {
	lastReturn := <-rwc.returnChan
	n = copy(p, lastReturn)
	return
}

func (rwc TestTunReadWriteCloser) Write(p []byte) (n int, err error) {
	rwc.lastWritten = p
	return len(p), nil
}

/* Return a slice of bytes to be read by the listeners on the tun interface. */
func (rwc TestTunReadWriteCloser) Return(p []byte) {
	rwc.returnChan <- p
}

func (rwc TestTunReadWriteCloser) Close() error {
	close(rwc.returnChan)
	rwc.lastWritten = nil
	return nil
}

var ttrwc TestTunReadWriteCloser
var tun = &water.Interface{
	ReadWriteCloser: ttrwc,
}

func TestBook(t *testing.T) {
	pOut := make(chan packet.Packet)
	pIn := make(chan packet.Packet)
	vpnnet := "10.0.0.0/24"

	// Define the dummy packets
	op1 := packet.Packet{
		Header: packet.PacketHeader{
			Iv:  []byte("CLIENT_1"),
			Sk:  []byte("SESS01"),
			Len: 20,
		},
		Data: []byte{
			0x54, 0x00, 0x14, 0x00, // [Ver4 + IHL=160bit] [DSCP+ECN] [LEN_LOW=20] [LEN_HIGH]
			0x00, 0x00, 0x00, 0x00, // [ID_LOW=0] [ID_HIGH=0] [Flags=0, Fragment = 0]
			0xff, 0x06, 0x00, 0x00, // [TTL=255] [PROTO=TCP] [CHECKSUM=0]
			0x0A, 0x00, 0x00, 0x01, // [SOURCE_IP=10.0.0.1]
			0xCA, 0xFE, 0xBA, 0xBE, // [DEST_IP=CAFEBABE]
			0x00, 0x00, 0x00, 0x00, // [DATA=0*4]
		},
	}

	op2 := packet.Packet{
		Header: packet.PacketHeader{
			Iv:  []byte("CLIENT_2"),
			Sk:  []byte("SESS02"),
			Len: 20,
		},
		Data: []byte{
			0x54, 0x00, 0x14, 0x00, // [Ver4 + IHL=160bit] [DSCP+ECN] [LEN_LOW=20] [LEN_HIGH]
			0x00, 0x00, 0x00, 0x00, // [ID_LOW=0] [ID_HIGH=0] [Flags=0, Fragment = 0]
			0xff, 0x06, 0x00, 0x00, // [TTL=255] [PROTO=TCP] [CHECKSUM=0]
			0x0A, 0x00, 0x00, 0x02, // [SOURCE_IP=10.0.0.2]
			0xCA, 0xFE, 0xBA, 0xBE, // [DEST_IP=CAFEBABE]
			0x00, 0x00, 0x00, 0x00, // [DATA=0*4]
		},
	}

	ip1 := []byte{
		0x54, 0x00, 0x14, 0x00, // [Ver4 + IHL=160bit] [DSCP+ECN] [LEN_LOW=20] [LEN_HIGH]
		0x00, 0x00, 0x00, 0x00, // [ID_LOW=0] [ID_HIGH=0] [Flags=0, Fragment = 0]
		0xff, 0x06, 0x00, 0x00, // [TTL=255] [PROTO=TCP] [CHECKSUM=0]
		0xCA, 0xFE, 0xBA, 0xBE, // [SOURCE_IP=CAFEBABE]
		0x0A, 0x00, 0x00, 0x01, // [DEST_IP=10.0.0.1]
		0x00, 0x00, 0x00, 0x00, // [DATA=0*4]
	}

	ip2 := []byte{
		0x54, 0x00, 0x14, 0x00, // [Ver4 + IHL=160bit] [DSCP+ECN] [LEN_LOW=20] [LEN_HIGH]
		0x00, 0x00, 0x00, 0x00, // [ID_LOW=0] [ID_HIGH=0] [Flags=0, Fragment = 0]
		0xff, 0x06, 0x00, 0x00, // [TTL=255] [PROTO=TCP] [CHECKSUM=0]
		0xCA, 0xFE, 0xBA, 0xBE, // [SOURCE_IP=CAFEBABE]
		0x0A, 0x00, 0x00, 0x02, // [DEST_IP=10.0.0.2]
		0x00, 0x00, 0x00, 0x00, // [DATA=0*4]
	}

	bs, err := newBookServer(pOut, pIn, vpnnet, tun)
	bs.start()

	if err != nil {
		t.Error("Failed to create new book server", err)
	}
	pOut <- op1

	bytesOut := ttrwc.lastWritten
	// TODO: Test the correctness of the bytes out here
	if len(bytesOut) == 0 {
		t.Error("Zero bytes out")
	}

	pOut <- op2

	bytesOut = ttrwc.lastWritten
	// TODO: Test the correctness of the second packet bytes out

	ttrwc.Return(ip1) // This should be returned to client 1

	packetBack := <-pIn

	if bytes.Compare(packetBack.Header.Sk, []byte("CLIENT_1")) != 0 {
		t.Error("First packet returned did not get CLIENT_1 session key")
	}

	ttrwc.Return(ip2) // This should be returned to client 2

	packetBack = <-pIn

	if bytes.Compare(packetBack.Header.Sk, []byte("CLIENT_2")) != 0 {
		t.Error("Second packet returned did not get CLIENT_2 session key")
	}
}
