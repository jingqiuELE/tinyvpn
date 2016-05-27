package session

import (
	"crypto/rand"
	"fmt"
)

const KeyLen = 6

type Key [KeyLen]byte
type Secret []byte

func NewKey() (*Key, error) {
	k := new(Key)
	_, err := rand.Read(k[:])
	if err != nil {
		fmt.Println("Error:", err)
		return k, err
	}
	return k, err
}
