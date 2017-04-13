package rsautil

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func Test_Rsautil(t *testing.T) {
	var publicKey rsa.PublicKey
	var privateKey rsa.PrivateKey
	//Assume that session_key is a 16-bytes string
	var session_key = GetRandomSessionKey()
	var rng = rand.Reader
	var encrypt_session_key []byte
	var decrypt_session_key []byte
	var err error

	GenKey()
	LoadKey("public.key", &publicKey)
	encrypt_session_key, err = rsa.EncryptPKCS1v15(rng, &publicKey, session_key)
	if err != nil {
		t.Error(err)
		return
	}

	LoadKey("private.key", &privateKey)
	decrypt_session_key = GetRandomSessionKey()
	err = rsa.DecryptPKCS1v15SessionKey(rng, &privateKey, encrypt_session_key,
		decrypt_session_key)
	if err != nil {
		t.Error(err)
		return
	}

	if bytes.Compare(session_key, decrypt_session_key) != 0 {
		t.Error("Decrypted session key is NOT the same as the plain text")
		return
	}
}
