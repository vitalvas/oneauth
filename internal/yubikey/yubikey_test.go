package yubikey

import (
	"testing"

	"github.com/go-piv/piv-go/v2/piv"
	"github.com/stretchr/testify/assert"
)

func TestYubikeyStruct(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		var y Yubikey
		assert.Nil(t, y.yk)
		assert.Equal(t, uint32(0), y.Serial)
	})

	t.Run("with serial", func(t *testing.T) {
		y := Yubikey{Serial: 12345678}
		assert.Equal(t, uint32(12345678), y.Serial)
		assert.Nil(t, y.yk)
	})
}

func TestYubikey_Close(t *testing.T) {
	t.Run("nil yk", func(t *testing.T) {
		y := &Yubikey{}
		err := y.Close()
		assert.NoError(t, err)
	})
}

func TestYubikey_ReOpen_NilYK_ZeroSerial(t *testing.T) {
	y := &Yubikey{Serial: 0}
	err := y.reOpen()
	// With no hardware, reOpen should fail trying to open by serial
	assert.Error(t, err)
}

func TestYubikey_ListKeys_NilYK(t *testing.T) {
	y := &Yubikey{Serial: 0}
	_, err := y.ListKeys()
	assert.Error(t, err)
}

func TestYubikey_Retries_NilYK(t *testing.T) {
	y := &Yubikey{Serial: 0}
	_, err := y.Retries()
	assert.Error(t, err)
}

func TestYubikey_VerifyPIN_NilYK(t *testing.T) {
	y := &Yubikey{Serial: 0}
	err := y.VerifyPIN("123456")
	assert.Error(t, err)
}

func TestYubikey_SetPIN_NilYK(t *testing.T) {
	y := &Yubikey{Serial: 0}
	err := y.SetPIN("123456", "654321")
	assert.Error(t, err)
}

func TestYubikey_SetPUK_NilYK(t *testing.T) {
	y := &Yubikey{Serial: 0}
	err := y.SetPUK("12345678", "87654321")
	assert.Error(t, err)
}

func TestYubikey_Unblock_NilYK(t *testing.T) {
	y := &Yubikey{Serial: 0}
	err := y.Unblock("12345678", "123456")
	assert.Error(t, err)
}

func TestYubikey_ResetToDefault_NilYK(t *testing.T) {
	y := &Yubikey{Serial: 0}
	err := y.ResetToDefault()
	assert.Error(t, err)
}

func TestYubikey_Reset_NilYK(t *testing.T) {
	y := &Yubikey{Serial: 0}
	err := y.Reset("123456", "12345678")
	assert.Error(t, err)
}

func TestYubikey_ResetMngmtKey_NilYK(t *testing.T) {
	y := &Yubikey{Serial: 0}
	err := y.ResetMngmtKey([]byte("123456789012345678901234"))
	assert.Error(t, err)
}

func TestCardStruct(t *testing.T) {
	t.Run("String method", func(t *testing.T) {
		c := Card{Name: "test", Serial: 123, Version: "5.4.3"}
		assert.Equal(t, "Yubikey #123", c.String())
	})

	t.Run("zero value String", func(t *testing.T) {
		c := Card{}
		assert.Equal(t, "Yubikey #0", c.String())
	})
}

func TestYubikey_PrivateKey_NilYK(t *testing.T) {
	y := &Yubikey{Serial: 0}
	_, err := y.PrivateKey(piv.SlotAuthentication, nil, piv.KeyAuth{})
	assert.Error(t, err)
}

func TestYubikey_OpenBySerial_NotFound(t *testing.T) {
	// Serial that won't match any connected device
	_, err := OpenBySerial(999999999)
	assert.Error(t, err)
}

func TestYubikey_Open_InvalidCard(t *testing.T) {
	card := Card{
		Name:    "non-existent-card",
		Serial:  12345,
		Version: "5.0.0",
	}
	_, err := Open(card)
	assert.Error(t, err)
}

func TestYubikey_Close_NilYK_Multiple(t *testing.T) {
	y := &Yubikey{}
	// Close should be safe to call multiple times on nil yk
	err := y.Close()
	assert.NoError(t, err)
	err = y.Close()
	assert.NoError(t, err)
}

func TestYubikey_ReOpen_NilYK_WithSerial(t *testing.T) {
	y := &Yubikey{Serial: 999999999}
	err := y.reOpen()
	// Should fail trying to open by serial with no hardware
	assert.Error(t, err)
}

func TestYubikey_ListKeys_NilYK_WithSlots(t *testing.T) {
	y := &Yubikey{Serial: 0}
	_, err := y.ListKeys(AllSSHSlots...)
	assert.Error(t, err)
}

func TestYubikey_GetActiveSlots_NilYK_EmptySlots(t *testing.T) {
	y := &Yubikey{Serial: 0}
	// Even with empty slots, it should fail on reOpen
	_, err := y.GetActiveSlots()
	assert.Error(t, err)
}
