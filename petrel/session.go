//session includes the type definition for a client's session.
//A session has an index, which maps to a randomly generated session secret.
package main

import (
	"crypto/rand"
	"fmt"
)

const IndexLen = 6
const SecretLen = 32

type Index [IndexLen]byte
type Secret [SecretLen]byte

func NewIndex() (*Index, error) {
	k := new(Index)
	_, err := rand.Read(k[:])
	if err != nil {
		fmt.Println("Error:", err)
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
