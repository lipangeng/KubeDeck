package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var (
	ErrInvalidKey        = errors.New("invalid encryption key")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

// Encryptor handles AES-GCM encryption/decryption
type Encryptor struct {
	key []byte
	aes cipher.Block
	gcm cipher.AEAD
}

// NewEncryptor creates a new encryptor with the given key
func NewEncryptor(key string) (*Encryptor, error) {
	keyBytes := []byte(key)

	// Ensure key is correct length (32 bytes for AES-256)
	if len(keyBytes) < 32 {
		// Pad key to 32 bytes
		padded := make([]byte, 32)
		copy(padded, keyBytes)
		keyBytes = padded
	} else if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}

	aesBlock, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, ErrInvalidKey
	}

	gcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, ErrInvalidKey
	}

	return &Encryptor{
		key: keyBytes,
		aes: aesBlock,
		gcm: gcm,
	}, nil
}

// Encrypt encrypts plaintext and returns base64-encoded ciphertext
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	nonceSize := e.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	return string(plaintext), nil
}

// GenerateKey generates a random 32-byte encryption key
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(key), nil
}
