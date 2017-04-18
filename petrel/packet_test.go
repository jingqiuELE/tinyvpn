package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"reflect"
	"testing"
)

const DATA_LEN = 200

func Test_EncodePacket(t *testing.T) {
	var err error

	p := new(Packet)
	_, err = rand.Read(p.Sk[:])
	if err != nil {
		t.Error(err)
		return
	}

	p.Data = make([]byte, DATA_LEN)
	n, err := rand.Read(p.Data)
	if err != nil {
		t.Error(err)
		return
	}
	p.Len = uint16(n)

	buf := new(bytes.Buffer)

	err = Encode(p, buf)
	if err != nil {
		t.Error(err)
		return
	}

	q, err := Decode(buf)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(p, q) {
		fmt.Println(p, reflect.TypeOf(p))
		fmt.Println(q, reflect.TypeOf(q))
		t.Error("Decoded packet not equal to the original one")
		return
	}

}
