package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

type Encryptor struct {
	aead cipher.AEAD
}

func New(keyBase64 string) (*Encryptor, error) {

	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("decode encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	return &Encryptor{aead: aead}, nil
}

func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	sealed := e.aead.Seal(nil, nonce, []byte(plaintext), nil)

	combined := make([]byte, 0, len(nonce)+len(sealed))
	combined = append(combined, nonce...)
	combined = append(combined, sealed...)

	return base64.StdEncoding.EncodeToString(combined), nil
}

func (e *Encryptor) Decrypt(encoded string) (string, error) {
	combined, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decrypt ciphertext: %w", err)
	}
	nonceSize := e.aead.NonceSize()
	if len(combined) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce := combined[:nonceSize]
	ciphertext := combined[nonceSize:]

	plaintext, err := e.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt ciphertext: %w", err)
	}
	return string(plaintext), nil
}
