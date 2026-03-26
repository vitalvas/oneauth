package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYubikeyChangePukCmd(t *testing.T) {
	t.Run("CommandMetadata", func(t *testing.T) {
		assert.Equal(t, "change-puk", yubikeyChangePukCmd.Name)
		assert.Equal(t, "Change the PUK of a YubiKey", yubikeyChangePukCmd.Usage)
		assert.NotNil(t, yubikeyChangePukCmd.Action)
		assert.NotNil(t, yubikeyChangePukCmd.Before)
	})

	t.Run("Flags", func(t *testing.T) {
		assert.NotEmpty(t, yubikeyChangePukCmd.Flags)

		flagNames := make(map[string]bool)
		for _, f := range yubikeyChangePukCmd.Flags {
			for _, name := range f.Names() {
				flagNames[name] = true
			}
		}

		assert.True(t, flagNames["serial"], "expected serial flag to exist")
	})
}
