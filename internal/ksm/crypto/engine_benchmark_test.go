package crypto

import (
	"testing"
)

func BenchmarkNewEngine(b *testing.B) {
	masterKey := "test-master-key-1234567890"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewEngine(masterKey)
	}
}

func BenchmarkEncryptAESKey(b *testing.B) {
	engine, _ := NewEngine("test-master-key-1234567890")
	keyID := "cccccccccccc"
	aesKey := []byte("1234567890123456")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.EncryptAESKey(keyID, aesKey)
	}
}

func BenchmarkDecryptAESKey(b *testing.B) {
	engine, _ := NewEngine("test-master-key-1234567890")
	keyID := "cccccccccccc"
	aesKey := []byte("1234567890123456")

	// Pre-encrypt data for benchmarking
	encrypted, _ := engine.EncryptAESKey(keyID, aesKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.DecryptAESKey(keyID, encrypted)
	}
}

func BenchmarkDeriveRowKey(b *testing.B) {
	engine, _ := NewEngine("test-master-key-1234567890")
	keyID := "cccccccccccc"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rowKey, _ := engine.deriveRowKey(keyID)
		clear(rowKey) // Clean up
	}
}

func BenchmarkDecryptYubikeyOTP(b *testing.B) {
	engine, _ := NewEngine("test-master-key-1234567890")
	otp := "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic"
	aesKey := []byte("1234567890123456")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.DecryptYubikeyOTP(otp, aesKey)
	}
}

func BenchmarkEncryptDecryptRoundtrip(b *testing.B) {
	engine, _ := NewEngine("test-master-key-1234567890")
	keyID := "cccccccccccc"
	aesKey := []byte("1234567890123456")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encrypted, _ := engine.EncryptAESKey(keyID, aesKey)
		decrypted, _ := engine.DecryptAESKey(keyID, encrypted)
		clear(decrypted) // Clean up
	}
}
