package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKeystore(t *testing.T) {
	t.Run("CreateNewKeystore", func(t *testing.T) {
		keystore := NewKeystore()
		require.NotNil(t, keystore)

		// Test basic functionality
		assert.Equal(t, 0, keystore.Len())
		assert.Empty(t, keystore.List())

		// Test RemoveAll doesn't panic
		assert.NotPanics(t, func() {
			keystore.RemoveAll()
		})
	})
}
