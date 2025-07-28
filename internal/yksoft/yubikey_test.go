package yksoft

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vitalvas/oneauth/internal/ykshared"
)

func TestSoftwareYubikey(t *testing.T) {
	t.Run("NewSoftwareYubikey", func(t *testing.T) {
		t.Run("default configuration", func(t *testing.T) {
			yk, err := NewSoftwareYubikey(nil)
			assert.NoError(t, err)
			assert.NotNil(t, yk)

			// Verify KeyID format
			assert.Len(t, yk.KeyID, 12)
			assert.NoError(t, ykshared.ValidateKeyIDFormat(yk.KeyID))

			// Verify PrivateID
			assert.Len(t, yk.PrivateID, 6)

			// Verify AES key
			assert.Len(t, yk.AESKey, 16)

			// Verify initial state
			assert.Equal(t, uint16(1), yk.Counter)
			assert.Equal(t, uint8(0), yk.SessionUse)
			assert.NotZero(t, yk.Timestamp)
			assert.NotZero(t, yk.RandomSeed)
		})

		t.Run("custom configuration", func(t *testing.T) {
			keyID := "cccccccccccc"
			privateID := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
			aesKey := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00}

			config := &Config{
				KeyID:     keyID,
				PrivateID: privateID,
				AESKey:    aesKey,
			}

			yk, err := NewSoftwareYubikey(config)
			assert.NoError(t, err)
			assert.NotNil(t, yk)

			assert.Equal(t, keyID, yk.KeyID)
			assert.Equal(t, privateID, yk.PrivateID)
			assert.Equal(t, aesKey, yk.AESKey)
		})

		t.Run("partial configurations", func(t *testing.T) {
			t.Run("empty KeyID triggers generation", func(t *testing.T) {
				config := &Config{KeyID: ""}
				yk, err := NewSoftwareYubikey(config)
				assert.NoError(t, err)
				assert.Len(t, yk.KeyID, 12)
				assert.NoError(t, ykshared.ValidateKeyIDFormat(yk.KeyID))
			})

			t.Run("only KeyID provided", func(t *testing.T) {
				keyID := "dddddddddddd"
				config := &Config{KeyID: keyID}
				yk, err := NewSoftwareYubikey(config)
				assert.NoError(t, err)
				assert.Equal(t, keyID, yk.KeyID)
				assert.Len(t, yk.PrivateID, 6)
				assert.Len(t, yk.AESKey, 16)
			})

			t.Run("only PrivateID provided", func(t *testing.T) {
				privateID := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
				config := &Config{PrivateID: privateID}
				yk, err := NewSoftwareYubikey(config)
				assert.NoError(t, err)
				assert.Equal(t, privateID, yk.PrivateID)
				assert.Len(t, yk.KeyID, 12)
				assert.Len(t, yk.AESKey, 16)
			})

			t.Run("only AESKey provided", func(t *testing.T) {
				aesKey := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10}
				config := &Config{AESKey: aesKey}
				yk, err := NewSoftwareYubikey(config)
				assert.NoError(t, err)
				assert.Equal(t, aesKey, yk.AESKey)
				assert.Len(t, yk.KeyID, 12)
				assert.Len(t, yk.PrivateID, 6)
			})
		})

		t.Run("invalid configurations", func(t *testing.T) {
			t.Run("invalid KeyID", func(t *testing.T) {
				tests := []struct {
					name  string
					keyID string
				}{
					{"too short", "invalid"},
					{"invalid modhex chars", "ccccccccccaX"},
					{"too long", "ccccccccccccc"},
				}

				for _, tt := range tests {
					t.Run(tt.name, func(t *testing.T) {
						config := &Config{KeyID: tt.keyID}
						_, err := NewSoftwareYubikey(config)
						assert.Error(t, err)
						assert.Contains(t, err.Error(), "invalid KeyID")
					})
				}
			})

			t.Run("invalid PrivateID", func(t *testing.T) {
				tests := []struct {
					name      string
					privateID []byte
				}{
					{"too short", []byte{0x01, 0x02}},
					{"too long", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}},
					{"empty slice", []byte{}},
				}

				for _, tt := range tests {
					t.Run(tt.name, func(t *testing.T) {
						config := &Config{PrivateID: tt.privateID}
						_, err := NewSoftwareYubikey(config)
						assert.Error(t, err)
						assert.Contains(t, err.Error(), "PrivateID must be exactly 6 bytes")
					})
				}
			})

			t.Run("invalid AES key", func(t *testing.T) {
				tests := []struct {
					name   string
					aesKey []byte
				}{
					{"too short", []byte{0x01, 0x02, 0x03}},
					{"too long", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11}},
					{"empty slice", []byte{}},
				}

				for _, tt := range tests {
					t.Run(tt.name, func(t *testing.T) {
						config := &Config{AESKey: tt.aesKey}
						_, err := NewSoftwareYubikey(config)
						assert.Error(t, err)
						assert.Contains(t, err.Error(), "AES key must be exactly 16 bytes")
					})
				}
			})
		})

		t.Run("initial values", func(t *testing.T) {
			yk, err := NewSoftwareYubikey(nil)
			assert.NoError(t, err)

			t.Run("counter starts at 1", func(t *testing.T) {
				assert.Equal(t, uint16(1), yk.Counter)
			})

			t.Run("session use starts at 0", func(t *testing.T) {
				assert.Equal(t, uint8(0), yk.SessionUse)
			})

			t.Run("creation time is set", func(t *testing.T) {
				assert.False(t, yk.Created.IsZero())
				assert.WithinDuration(t, time.Now(), yk.Created, time.Second)
			})

			t.Run("timestamp is valid 24-bit value", func(t *testing.T) {
				assert.NotZero(t, yk.Timestamp)
				assert.LessOrEqual(t, yk.Timestamp, uint32(0xFFFFFF))
			})

			t.Run("random seed is set", func(t *testing.T) {
				assert.NotZero(t, yk.RandomSeed)
			})
		})
	})

	t.Run("GenerateOTP", func(t *testing.T) {
		t.Run("basic functionality", func(t *testing.T) {
			yk, err := NewSoftwareYubikey(nil)
			assert.NoError(t, err)

			result, err := yk.GenerateOTP()
			assert.NoError(t, err)
			assert.NotNil(t, result)

			t.Run("OTP format validation", func(t *testing.T) {
				assert.Len(t, result.OTP, 44)
				assert.True(t, ykshared.IsValidModhex(result.OTP))
				assert.Equal(t, yk.KeyID, result.OTP[:12])

				encryptedPart := result.OTP[12:]
				assert.Len(t, encryptedPart, 32)
				assert.True(t, ykshared.IsValidModhex(encryptedPart))
			})

			t.Run("result metadata", func(t *testing.T) {
				assert.Equal(t, yk.Counter, result.Counter)
				assert.Equal(t, yk.SessionUse, result.SessionUse)
				assert.Equal(t, uint8(1), result.SessionUse) // First use should be 1
				assert.Equal(t, yk.Timestamp, result.Timestamp)
				assert.NotZero(t, result.CRC)
			})
		})

		t.Run("session management", func(t *testing.T) {
			yk, err := NewSoftwareYubikey(nil)
			assert.NoError(t, err)

			t.Run("multiple OTPs increment session use", func(t *testing.T) {
				results := make([]*OTPResult, 3)
				for i := 0; i < 3; i++ {
					result, err := yk.GenerateOTP()
					assert.NoError(t, err)
					results[i] = result
				}

				// Verify session use increments
				assert.Equal(t, uint8(1), results[0].SessionUse)
				assert.Equal(t, uint8(2), results[1].SessionUse)
				assert.Equal(t, uint8(3), results[2].SessionUse)

				// Verify all have same counter
				assert.Equal(t, results[0].Counter, results[1].Counter)
				assert.Equal(t, results[1].Counter, results[2].Counter)

				// Verify OTPs are different
				assert.NotEqual(t, results[0].OTP, results[1].OTP)
				assert.NotEqual(t, results[1].OTP, results[2].OTP)
			})

			t.Run("session use overflow", func(t *testing.T) {
				yk.SessionUse = 254

				result1, err := yk.GenerateOTP()
				assert.NoError(t, err)
				assert.Equal(t, uint8(255), result1.SessionUse)

				result2, err := yk.GenerateOTP()
				assert.NoError(t, err)
				assert.Equal(t, uint8(0), result2.SessionUse) // Wraps around
			})
		})

		t.Run("counter management", func(t *testing.T) {
			yk, err := NewSoftwareYubikey(nil)
			assert.NoError(t, err)

			t.Run("counter increment resets session use", func(t *testing.T) {
				// Generate OTPs to increment session use
				_, err := yk.GenerateOTP()
				assert.NoError(t, err)
				assert.Greater(t, yk.SessionUse, uint8(0))

				initialCounter := yk.Counter
				yk.IncrementCounter()

				result, err := yk.GenerateOTP()
				assert.NoError(t, err)
				assert.Equal(t, uint8(1), result.SessionUse) // Reset to 1
				assert.Equal(t, initialCounter+1, result.Counter)
			})

			t.Run("counter overflow", func(t *testing.T) {
				yk.Counter = 65535

				result1, err := yk.GenerateOTP()
				assert.NoError(t, err)
				assert.Equal(t, uint16(65535), result1.Counter)

				yk.IncrementCounter()
				result2, err := yk.GenerateOTP()
				assert.NoError(t, err)
				assert.Equal(t, uint16(0), result2.Counter) // Wraps to 0
			})
		})

		t.Run("consistency and uniqueness", func(t *testing.T) {
			t.Run("same config produces different instances", func(t *testing.T) {
				config := &Config{KeyID: "cccccccccccc"}

				yk1, err := NewSoftwareYubikey(config)
				assert.NoError(t, err)
				yk2, err := NewSoftwareYubikey(config)
				assert.NoError(t, err)

				// Same KeyID but different secrets
				assert.Equal(t, yk1.KeyID, yk2.KeyID)
				assert.NotEqual(t, yk1.PrivateID, yk2.PrivateID)
				assert.NotEqual(t, yk1.AESKey, yk2.AESKey)
			})

			t.Run("OTPs from same key are unique", func(t *testing.T) {
				yk, err := NewSoftwareYubikey(nil)
				assert.NoError(t, err)

				result1, err := yk.GenerateOTP()
				assert.NoError(t, err)
				result2, err := yk.GenerateOTP()
				assert.NoError(t, err)

				assert.NotEqual(t, result1.OTP, result2.OTP)
				assert.NotEqual(t, result1.SessionUse, result2.SessionUse)
			})

			t.Run("timestamp updates", func(t *testing.T) {
				yk, err := NewSoftwareYubikey(nil)
				assert.NoError(t, err)

				initialTimestamp := yk.Timestamp
				time.Sleep(time.Millisecond * 10)

				result, err := yk.GenerateOTP()
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, result.Timestamp, initialTimestamp)
			})
		})

		t.Run("compatibility", func(t *testing.T) {
			t.Run("OTP passes validation", func(t *testing.T) {
				keyID := "cccccccccccc"
				privateID := []byte{0x87, 0x92, 0xeb, 0xfe, 0x26, 0xcc}
				aesKey := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00}

				config := &Config{
					KeyID:     keyID,
					PrivateID: privateID,
					AESKey:    aesKey,
				}

				yk, err := NewSoftwareYubikey(config)
				assert.NoError(t, err)

				result, err := yk.GenerateOTP()
				assert.NoError(t, err)

				// Should pass ykshared validation
				_, err = ykshared.ValidateOTP(result.OTP)
				assert.NoError(t, err)
			})
		})
	})

	t.Run("IncrementCounter", func(t *testing.T) {
		yk, err := NewSoftwareYubikey(nil)
		assert.NoError(t, err)

		t.Run("counter increments correctly", func(t *testing.T) {
			initialCounter := yk.Counter
			yk.IncrementCounter()
			assert.Equal(t, initialCounter+1, yk.Counter)
		})

		t.Run("session use resets to 0", func(t *testing.T) {
			// Generate OTPs to increment session use
			_, err := yk.GenerateOTP()
			assert.NoError(t, err)
			_, err = yk.GenerateOTP()
			assert.NoError(t, err)
			assert.Greater(t, yk.SessionUse, uint8(0))

			yk.IncrementCounter()
			assert.Equal(t, uint8(0), yk.SessionUse)
		})

		t.Run("multiple increments", func(t *testing.T) {
			initialCounter := yk.Counter
			yk.IncrementCounter()
			yk.IncrementCounter()
			yk.IncrementCounter()
			assert.Equal(t, initialCounter+3, yk.Counter)
			assert.Equal(t, uint8(0), yk.SessionUse)
		})
	})

	t.Run("Getter methods", func(t *testing.T) {
		keyID := "cccccccccccc"
		privateID := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
		aesKey := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00}

		config := &Config{
			KeyID:     keyID,
			PrivateID: privateID,
			AESKey:    aesKey,
		}

		yk, err := NewSoftwareYubikey(config)
		assert.NoError(t, err)

		t.Run("GetKeyID", func(t *testing.T) {
			assert.Equal(t, keyID, yk.GetKeyID())
		})

		t.Run("GetAESKey returns copy", func(t *testing.T) {
			key1 := yk.GetAESKey()
			key2 := yk.GetAESKey()
			assert.Equal(t, aesKey, key1)
			assert.Equal(t, key1, key2)

			// Modify returned key shouldn't affect original
			key1[0] = 0xFF
			key2 = yk.GetAESKey()
			assert.NotEqual(t, key1[0], key2[0])
		})

		t.Run("GetPrivateID returns copy", func(t *testing.T) {
			id1 := yk.GetPrivateID()
			id2 := yk.GetPrivateID()
			assert.Equal(t, privateID, id1)
			assert.Equal(t, id1, id2)

			// Modify returned ID shouldn't affect original
			id1[0] = 0xFF
			id2 = yk.GetPrivateID()
			assert.NotEqual(t, id1[0], id2[0])
		})
	})

	t.Run("Security", func(t *testing.T) {
		t.Run("configuration isolation", func(t *testing.T) {
			privateID := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
			aesKey := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00}

			config := &Config{
				PrivateID: privateID,
				AESKey:    aesKey,
			}

			yk, err := NewSoftwareYubikey(config)
			assert.NoError(t, err)

			// Modify original slices
			privateID[0] = 0xFF
			aesKey[0] = 0xFF

			// YubiKey should not be affected
			assert.NotEqual(t, privateID[0], yk.PrivateID[0])
			assert.NotEqual(t, aesKey[0], yk.AESKey[0])
		})

		t.Run("memory safety", func(t *testing.T) {
			yk, err := NewSoftwareYubikey(nil)
			assert.NoError(t, err)

			// Test GetAESKey returns copies
			key1 := yk.GetAESKey()
			key2 := yk.GetAESKey()
			assert.Equal(t, key1, key2)

			// Modify one copy shouldn't affect original
			originalKeyByte := key1[0]
			key1[0] = 0xFF
			key2 = yk.GetAESKey()
			assert.Equal(t, originalKeyByte, key2[0]) // Should get original value back
			assert.NotEqual(t, key1[0], key2[0])

			// Test GetPrivateID returns copies
			id1 := yk.GetPrivateID()
			id2 := yk.GetPrivateID()
			assert.Equal(t, id1, id2)

			// Modify one copy shouldn't affect original
			originalIDByte := id1[0]
			id1[0] = 0xFF
			id2 = yk.GetPrivateID()
			assert.Equal(t, originalIDByte, id2[0]) // Should get original value back
			assert.NotEqual(t, id1[0], id2[0])
		})
	})
}
