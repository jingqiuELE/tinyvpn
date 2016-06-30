package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"packet"
)

func Encrypt(key []byte, ivSize int, data []byte) (encryptedData, iv []byte, err error) {
	// Prepare the cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	// Create IV of required size
	ivSpace := aes.BlockSize
	if ivSize > aes.BlockSize {
		ivSpace = ivSize
	}
	realIv := make([]byte, ivSpace)
	iv = realIv[:ivSize]
	realIv = realIv[:aes.BlockSize]
	_, err = io.ReadFull(rand.Reader, iv)

	// Create CTR mode cipher
	encryptedData = make([]byte, len(data))
	ctr := cipher.NewCTR(block, realIv)
	ctr.XORKeyStream(encryptedData, data)

	return
}

func Decrypt(key, iv, encryptedData []byte) (data []byte, err error) {
	// Prepare the cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	// Prepare IV
	realIv := make([]byte, aes.BlockSize)
	copy(realIv, iv)

	// Create CTR mode cipher
	data = make([]byte, len(encryptedData))
	ctr := cipher.NewCTR(block, realIv)
	ctr.XORKeyStream(data, encryptedData)

	return
}

func EncryptPacket(p *packet.Packet, key []byte) error {
	encData, iv, err := Encrypt(key, packet.IvLen, p.Data)
	if err != nil {
		return err
	}

	copy(p.Header.Iv[:], iv)
	p.Data = encData

	return nil
}

func DecryptPacket(p *packet.Packet, key []byte) error {
	decData, err := Decrypt(key, p.Header.Iv[:], p.Data)
	if err != nil {
		return err
	}
	p.Data = decData
	return nil
}
