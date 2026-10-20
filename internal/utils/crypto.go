package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var (
	encryptionKey     []byte
	encryptionKeyErr  error
	encryptionKeyOnce sync.Once
)

/**************************************************************************************
* dataEncryptionKey returns the 32-byte AES key from the DATA_ENCRYPTION_KEY env var
* (base64-encoded). Generate one with: openssl rand -base64 32
* Loaded once per process; a missing or malformed key is a hard error — we never
* fall back to plaintext silently.
**************************************************************************************/
func dataEncryptionKey() ([]byte, error) {
	encryptionKeyOnce.Do(func() {
		raw := strings.TrimSpace(os.Getenv("DATA_ENCRYPTION_KEY"))

		if raw == "" {
			encryptionKeyErr = fmt.Errorf("DATA_ENCRYPTION_KEY is not set (generate one with: openssl rand -base64 32)")
			return
		}

		key, err := base64.StdEncoding.DecodeString(raw)
		if err != nil || len(key) != 32 {
			encryptionKeyErr = fmt.Errorf("DATA_ENCRYPTION_KEY must be 32 bytes, base64-encoded (generate: openssl rand -base64 32)")
			return
		}

		encryptionKey = key
	})

	return encryptionKey, encryptionKeyErr
}

/**************************************************************************************
* EncryptString encrypts plaintext with AES-256-GCM and returns
* base64(nonce || ciphertext). A fresh random nonce per call means encrypting the
* same value twice produces different ciphertexts.
*
* status: ✅
**************************************************************************************/
func EncryptString(plaintext string) (string, error) {
	key, err := dataEncryptionKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("could not create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("could not create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("could not generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

/**************************************************************************************
* DecryptString reverses EncryptString. Authentication failure (wrong key or
* tampered ciphertext) is an error, never a silent plaintext return.
*
* status: ✅
**************************************************************************************/
func DecryptString(encoded string) (string, error) {
	key, err := dataEncryptionKey()
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("could not decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("could not create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("could not create GCM: %w", err)
	}

	if len(data) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("could not decrypt (wrong key or corrupted data): %w", err)
	}

	return string(plaintext), nil
}
