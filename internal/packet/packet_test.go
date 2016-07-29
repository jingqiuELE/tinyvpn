package packet

import (
	"crypto/rand"
	"fmt"
	"github.com/jingqiuELE/tinyvpn/internal/session"
	"reflect"
	"testing"
)

const DATA_LEN = 200

func Test_MarshalPacket(t *testing.T) {
	var err error

	p := NewPacket()

	sk, err := session.NewIndex()
	if err != nil {
		t.Error(err)
		return
	}
	p.Header.Sk = *sk

	data := make([]byte, DATA_LEN)
	_, err = rand.Read(data)
	if err != nil {
		t.Error(err)
		return
	}
	p.SetData(data)

	buf, err := MarshalToSlice(*p)
	if err != nil {
		t.Error(err)
		return
	}

	q, err := UnmarshalFromSlice(buf)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(*p, q) {
		fmt.Println(p)
		fmt.Println(q)
		t.Error("unmarshaled packet not equal to the original one")
		return
	}

}
