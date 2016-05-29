package petrel

import (
	"bytes"
	"fmt"
	"packet"
	"testing"
)

func Test_Encrypt(t *testing.T) {
	key := make([]byte, 32)
	copy(key, []byte("Secret"))
	plain := []byte("Tiger, Tiger, Burining bright in the forest of the night.")
	fmt.Printf("Plain text : %v\n", plain)

	encrypted, iv, err := Encrypt(key, 8, plain)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("Encrypted text : %v\n", encrypted)

	if bytes.Compare(plain, encrypted) == 0 {
		t.Error("Encrypted text is the same as the plain text")
		return
	}

	plain2, err := Decrypt(key, iv, encrypted)
	fmt.Printf("Decrypted text : %v\n", plain2)

	if bytes.Compare(plain, plain2) != 0 {
		t.Error("Decrypted text is NOT the same as the plain text")
		return
	}
}

func Test_EncryptPacket(t *testing.T) {
	key := make([]byte, 32)
	copy(key, []byte("Secret"))

	const testText = "Some plain text to be encrypted and decrypted"
	p := packet.Packet{
		Header: packet.PacketHeader{
			Sk: []byte("SESSIN"),
		},
		Data: []byte(testText),
	}

	err := EncryptPacket(&p, key)
	if err != nil {
		t.Error("Failed to encrypt the packet.", err)
		return
	}

	if p.Header.Iv == nil {
		t.Error("IV was not generated.")
		return
	}

	if bytes.Compare(p.Data, []byte(testText)) == 0 {
		t.Error("Encrypted data is still the same.")
		return
	}

	err = DecryptPacket(&p, key)
	if err != nil {
		t.Error("Failed to Decrypt the packet.", err)
		return
	}

	if bytes.Compare(p.Data, []byte(testText)) != 0 {
		t.Error("Decrypted data failed to match the original text")
		return
	}

	fmt.Printf("Successfully encrypt and decrypt a packet. Generated IV is %v\n", p.Header.Iv)
}
