package main

import (
	"bytes"
	"testing"
	"time"

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

func TestBook(t *testing.T) {
	pOut := make(chan *Packet)
	pIn := make(chan *Packet)
	vpnnet := "10.0.0.0/24"

	var ttrwc TestTunReadWriteCloser
	ttrwc.returnChan = make(chan []byte)
	var tun = &water.Interface{
		ReadWriteCloser: ttrwc,
	}
	// Define the dummy packets
	op1 := Packet{
		Iv: Iv{'C', 'L', 'I', 'E', 'N', 'T', '_', '1'},
		Sk: Index{'S', 'E', 'S', 'S', '0', '1'},
		Data: []byte{
			0x54, 0x00, 0x14, 0x00, // [Ver4 + IHL=160bit] [DSCP+ECN] [LEN_LOW=20] [LEN_HIGH]
			0x00, 0x00, 0x00, 0x00, // [ID_LOW=0] [ID_HIGH=0] [Flags=0, Fragment = 0]
			0xff, 0x06, 0x00, 0x00, // [TTL=255] [PROTO=TCP] [CHECKSUM=0]
			0x0A, 0x00, 0x00, 0x01, // [SOURCE_IP=10.0.0.1]
			0xCA, 0xFE, 0xBA, 0xBE, // [DEST_IP=CAFEBABE]
			0x00, 0x00, 0x00, 0x00, // [DATA=0*4]
		},
	}

	op2 := Packet{
		Iv: Iv{'C', 'L', 'I', 'E', 'N', 'T', '_', '2'},
		Sk: Index{'S', 'E', 'S', 'S', '0', '2'},
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
	go bs.start()
	if err != nil {
		t.Error("Failed to create new book server", err)
	}
	pIn <- &op1
	pIn <- &op2

	// FIXME: using sleep to fix problem of returning faster than send, this might not always work!
	time.Sleep(time.Millisecond * 100)

	ttrwc.Return(ip1) // This should be returned to client 1

	packetBack := <-pOut
	backSessionKey := [IndexLen]byte(packetBack.Sk)
	if bytes.Compare(backSessionKey[:], []byte("SESS01")) != 0 {
		t.Error("First packet returned did not get CLIENT_1 session key")
	}

	ttrwc.Return(ip2) // This should be returned to client 2

	packetBack = <-pOut

	backSessionKey = [IndexLen]byte(packetBack.Sk)
	if bytes.Compare(backSessionKey[:], []byte("SESS02")) != 0 {
		t.Error("Second packet returned did not get CLIENT_2 session key")
	}
}
