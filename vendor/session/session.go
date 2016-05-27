package session

import (
	"crypto/rand"
	"fmt"
)

const IndexLen = 6

type Index [IndexLen]byte
type Secret []byte

func MakeIndex() (Index, error) {
	m := new(Index)
	_, err := rand.Read(m[:])
	if err != nil {
		fmt.Println("Error:", err)
		return *m, err
	}
	return *m, err
}
