package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

const (
	keySize   = 32
	nonceSize = 12
)

type Engine struct {
	masterKey []byte
}

func NewEngine(masterKey string) (*Engine, error) {
	if len(masterKey) == 0 {
		return nil, fmt.Errorf("master key cannot be empty")
	}

	key := sha256.Sum256([]byte(masterKey))
	return &Engine{
		masterKey: key[:],
	}, nil
}

func (e *Engine) deriveRowKey(keyID string) ([]byte, error) {
	hkdfReader := hkdf.New(sha256.New, e.masterKey, []byte("ksm-salt"), []byte(keyID))

	rowKey := make([]byte, keySize)
	if _, err := io.ReadFull(hkdfReader, rowKey); err != nil {
		return nil, fmt.Errorf("failed to derive row key: %w", err)
	}

	return rowKey, nil
}

func (e *Engine) EncryptAESKey(keyID string, aesKey []byte) (string, error) {
	if len(aesKey) != 16 {
		return "", fmt.Errorf("AES key must be 16 bytes")
	}

	rowKey, err := e.deriveRowKey(keyID)
	if err != nil {
		return "", fmt.Errorf("failed to derive row key: %w", err)
	}
	defer clear(rowKey)

	block, err := aes.NewCipher(rowKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	aad := []byte(keyID)
	ciphertext := gcm.Seal(nil, nonce, aesKey, aad)

	encryptedData := make([]byte, 0, len(nonce)+len(ciphertext))
	encryptedData = append(encryptedData, nonce...)
	encryptedData = append(encryptedData, ciphertext...)
	return base64.RawURLEncoding.EncodeToString(encryptedData), nil
}

func (e *Engine) DecryptAESKey(keyID string, encryptedData string) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	if len(data) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	rowKey, err := e.deriveRowKey(keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive row key: %w", err)
	}
	defer clear(rowKey)

	block, err := aes.NewCipher(rowKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]
	aad := []byte(keyID)

	plaintext, err := gcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}
