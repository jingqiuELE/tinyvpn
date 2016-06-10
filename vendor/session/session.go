package session

import (
	"crypto/rand"
	"fmt"
)

const KeyLen = 6

type Key [KeyLen]byte
type Secret [32]byte

func NewKey() (*Key, error) {
	k := new(Key)
	_, err := rand.Read(k[:])
	if err != nil {
		fmt.Println("Error:", err)
		return k, err
	}
	return k, err
}

func NewSecret() (*Secret, error) {
	s := new(Secret)
	_, err := rand.Read(s[:])
	if err != nil {
		fmt.Println("Error:", err)
		return s, err
	}
	return s, err
}
