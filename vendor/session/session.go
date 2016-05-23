package session

import (
	"crypto/rand"
	"fmt"
)

type SessionKey [6]byte
type Secret []byte

func NewSessionKey() (*SessionKey, error) {
	m := new(SessionKey)
	_, err := rand.Read(m[:])
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	return m, err
}
