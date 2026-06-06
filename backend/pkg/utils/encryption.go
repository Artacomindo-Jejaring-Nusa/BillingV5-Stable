package utils

import (
	"errors"
	"log"
	"strings"

	"github.com/fernet/fernet-go"
)

type EncryptionService struct {
	keys []*fernet.Key
}

var GlobalEncryptionService *EncryptionService

// InitEncryptionService initializes the global encryption service with the provided key.
func InitEncryptionService(keyStr string) error {
	if keyStr == "" {
		return errors.New("ENCRYPTION_KEY cannot be empty")
	}

	keys, err := fernet.DecodeKeys(keyStr)
	if err != nil {
		return err
	}

	GlobalEncryptionService = &EncryptionService{
		keys: keys,
	}
	return nil
}

// IsEncrypted check if a value is encrypted (starts with 'gAAAAA')
func (s *EncryptionService) IsEncrypted(value string) bool {
	return strings.HasPrefix(value, "gAAAAA")
}

// Encrypt encrypts a plaintext string into Fernet ciphertext
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" || s.IsEncrypted(plaintext) {
		return plaintext, nil
	}

	if len(s.keys) == 0 {
		return "", errors.New("encryption service not initialized or has no keys")
	}

	tok, err := fernet.EncryptAndSign([]byte(plaintext), s.keys[0])
	if err != nil {
		return "", err
	}

	return string(tok), nil
}

// Decrypt decrypts a Fernet ciphertext back to plaintext
func (s *EncryptionService) Decrypt(ciphertext string) string {
	if ciphertext == "" || !s.IsEncrypted(ciphertext) {
		return ciphertext
	}

	if len(s.keys) == 0 {
		log.Println("FATAL: Encryption service not initialized or has no keys")
		return ciphertext
	}

	// We pass 0 as TTL to disable expiration check, matching Python's default behavior
	decrypted := fernet.VerifyAndDecrypt([]byte(ciphertext), 0, s.keys)
	if decrypted == nil {
		log.Printf("FATAL: Failed to decrypt data: signature is invalid or key is wrong")
		return ciphertext
	}

	return string(decrypted)
}

// Helper functions for easy import/usage

func Encrypt(plaintext string) string {
	if GlobalEncryptionService == nil {
		return plaintext
	}
	val, err := GlobalEncryptionService.Encrypt(plaintext)
	if err != nil {
		log.Printf("Error encrypting data: %v", err)
		return plaintext
	}
	return val
}

func Decrypt(ciphertext string) string {
	if GlobalEncryptionService == nil {
		return ciphertext
	}
	return GlobalEncryptionService.Decrypt(ciphertext)
}
