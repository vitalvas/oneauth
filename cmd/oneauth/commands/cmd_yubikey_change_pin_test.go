package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYubikeyChangePinCmd(t *testing.T) {
	t.Run("CommandMetadata", func(t *testing.T) {
		assert.Equal(t, "change-pin", yubikeyChangePinCmd.Name)
		assert.Equal(t, "Change the PIN of a YubiKey", yubikeyChangePinCmd.Usage)
		assert.NotNil(t, yubikeyChangePinCmd.Action)
		assert.NotNil(t, yubikeyChangePinCmd.Before)
	})

	t.Run("Flags", func(t *testing.T) {
		assert.NotEmpty(t, yubikeyChangePinCmd.Flags)

		flagNames := make(map[string]bool)
		for _, f := range yubikeyChangePinCmd.Flags {
			for _, name := range f.Names() {
				flagNames[name] = true
			}
		}

		assert.True(t, flagNames["serial"], "expected serial flag to exist")
	})
}
