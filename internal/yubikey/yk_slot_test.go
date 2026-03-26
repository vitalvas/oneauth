package yubikey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetActiveSlots_NilYubikey(t *testing.T) {
	// Yubikey with nil yk should fail on reOpen inside ListKeys
	y := &Yubikey{Serial: 0}
	_, err := y.GetActiveSlots(AllSlots...)
	assert.Error(t, err)
}

func TestSlotVariables(t *testing.T) {
	t.Run("default slot key IDs", func(t *testing.T) {
		assert.Equal(t, uint32(0x95), SlotKeyRSAID)
		assert.Equal(t, uint32(0x94), SlotKeyECDSAID)
	})

	t.Run("SlotKeyRSA matches ID", func(t *testing.T) {
		assert.Equal(t, SlotKeyRSAID, SlotKeyRSA.PIVSlot.Key)
	})

	t.Run("SlotKeyECDSA matches ID", func(t *testing.T) {
		assert.Equal(t, SlotKeyECDSAID, SlotKeyECDSA.PIVSlot.Key)
	})

	t.Run("AllSSHSlots contains expected slots", func(t *testing.T) {
		assert.Len(t, AllSSHSlots, 2)
		assert.Equal(t, SlotKeyRSA, AllSSHSlots[0])
		assert.Equal(t, SlotKeyECDSA, AllSSHSlots[1])
	})

	t.Run("AllSlots is non-empty", func(t *testing.T) {
		assert.True(t, len(AllSlots) > 0)
	})

	t.Run("PIVSlots matches AllSlots length", func(t *testing.T) {
		assert.Equal(t, len(AllSlots), len(PIVSlots))
	})

	t.Run("PIVSlots strings are non-empty", func(t *testing.T) {
		for _, s := range PIVSlots {
			assert.NotEmpty(t, s)
		}
	})
}
