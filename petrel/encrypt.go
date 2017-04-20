package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

func getGCM(key []byte, ivSize int) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, ivSize)
	if err != nil {
		return nil, err
	}

	return gcm, nil
}

func Encrypt(key []byte, ivSize int, data []byte) (encryptedData, iv []byte, err error) {
	gcm, err := getGCM(key, ivSize)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, nil, err
	}

	return gcm.Seal(nil, nonce, data, nil), nonce[:ivSize], nil
}

func Decrypt(key, iv, encryptedData []byte) (data []byte, err error) {
	gcm, err := getGCM(key, len(iv))
	if err != nil {
		return nil, err
	}

	if len(encryptedData) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil, iv, encryptedData, nil)
}

func EncryptPacket(p *Packet, key []byte) error {
	encrypted, iv, err := Encrypt(key, IvLen, p.Data)
	if err != nil {
		return err
	}

	p.EncryptedData = encrypted
	p.Iv = iv

	return nil
}

func DecryptPacket(p *Packet, key []byte) error {
	data, err := Decrypt(key, p.Iv[:], p.EncryptedData)
	if err != nil {
		return err
	}
	p.Data = data
	return nil
}

// TODO: Create multiple go routines for encryption/decryption, possibly dynamaically scale
func encryptPackets(from <-chan *Packet, to chan<- *Packet, ss SecretSource) {
	for {
		p := <-from
		key, ok := ss.getSecret(p.Sk)
		if !ok {
			log.Error("Cannot find secret for session:", p.Sk)
			continue
		}
		EncryptPacket(p, key[:])
		to <- p
	}
}

func decryptPackets(from <-chan *Packet, to chan<- *Packet, ss SecretSource) {
	for {
		p := <-from
		key, ok := ss.getSecret(p.Sk)
		if !ok {
			log.Error("Cannot find secret for session:", p.Sk)
			continue
		}
		DecryptPacket(p, key[:])
		to <- p
	}
}
