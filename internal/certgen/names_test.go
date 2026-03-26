package certgen

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseExtraNames_Names(t *testing.T) {
	t.Run("all fields present", func(t *testing.T) {
		names := []pkix.AttributeTypeAndValue{
			{Type: ExtNameTokenID, Value: "token-123"},
			{Type: ExtNameTouchPolicy, Value: "always"},
			{Type: ExtNamePinPolicy, Value: "once"},
		}

		result, err := ParseExtraNames(names)
		assert.NoError(t, err)
		assert.Equal(t, "token-123", result.TokenID)
		assert.Equal(t, "always", result.TouchPolicy)
		assert.Equal(t, "once", result.PinPolicy)
	})

	t.Run("empty names slice", func(t *testing.T) {
		result, err := ParseExtraNames(nil)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.TokenID)
		assert.Empty(t, result.TouchPolicy)
		assert.Empty(t, result.PinPolicy)
	})

	t.Run("only token ID present", func(t *testing.T) {
		names := []pkix.AttributeTypeAndValue{
			{Type: ExtNameTokenID, Value: "my-token"},
		}

		result, err := ParseExtraNames(names)
		assert.NoError(t, err)
		assert.Equal(t, "my-token", result.TokenID)
		assert.Empty(t, result.TouchPolicy)
		assert.Empty(t, result.PinPolicy)
	})

	t.Run("only touch policy present", func(t *testing.T) {
		names := []pkix.AttributeTypeAndValue{
			{Type: ExtNameTouchPolicy, Value: "never"},
		}

		result, err := ParseExtraNames(names)
		assert.NoError(t, err)
		assert.Empty(t, result.TokenID)
		assert.Equal(t, "never", result.TouchPolicy)
		assert.Empty(t, result.PinPolicy)
	})

	t.Run("only pin policy present", func(t *testing.T) {
		names := []pkix.AttributeTypeAndValue{
			{Type: ExtNamePinPolicy, Value: "always"},
		}

		result, err := ParseExtraNames(names)
		assert.NoError(t, err)
		assert.Empty(t, result.TokenID)
		assert.Empty(t, result.TouchPolicy)
		assert.Equal(t, "always", result.PinPolicy)
	})

	t.Run("unknown OID is ignored", func(t *testing.T) {
		unknownOID := asn1.ObjectIdentifier([]int{1, 2, 3, 4, 5})
		names := []pkix.AttributeTypeAndValue{
			{Type: unknownOID, Value: "ignored-value"},
			{Type: ExtNameTokenID, Value: "token-456"},
		}

		result, err := ParseExtraNames(names)
		assert.NoError(t, err)
		assert.Equal(t, "token-456", result.TokenID)
	})

	t.Run("token ID with wrong value type returns error", func(t *testing.T) {
		names := []pkix.AttributeTypeAndValue{
			{Type: ExtNameTokenID, Value: 12345},
		}

		result, err := ParseExtraNames(names)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unexpected value type for token ID")
	})

	t.Run("touch policy with wrong value type returns error", func(t *testing.T) {
		names := []pkix.AttributeTypeAndValue{
			{Type: ExtNameTouchPolicy, Value: true},
		}

		result, err := ParseExtraNames(names)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unexpected value type for touch policy")
	})

	t.Run("pin policy with wrong value type returns error", func(t *testing.T) {
		names := []pkix.AttributeTypeAndValue{
			{Type: ExtNamePinPolicy, Value: []byte("bytes")},
		}

		result, err := ParseExtraNames(names)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unexpected value type for pin policy")
	})

	t.Run("duplicate fields uses last value", func(t *testing.T) {
		names := []pkix.AttributeTypeAndValue{
			{Type: ExtNameTokenID, Value: "first"},
			{Type: ExtNameTokenID, Value: "second"},
		}

		result, err := ParseExtraNames(names)
		assert.NoError(t, err)
		assert.Equal(t, "second", result.TokenID)
	})

	t.Run("empty string values are valid", func(t *testing.T) {
		names := []pkix.AttributeTypeAndValue{
			{Type: ExtNameTokenID, Value: ""},
			{Type: ExtNameTouchPolicy, Value: ""},
			{Type: ExtNamePinPolicy, Value: ""},
		}

		result, err := ParseExtraNames(names)
		assert.NoError(t, err)
		assert.Empty(t, result.TokenID)
		assert.Empty(t, result.TouchPolicy)
		assert.Empty(t, result.PinPolicy)
	})
}

func TestExtNameOIDs_Names(t *testing.T) {
	t.Run("OIDs are distinct", func(t *testing.T) {
		assert.False(t, ExtNameTokenID.Equal(ExtNameTouchPolicy))
		assert.False(t, ExtNameTokenID.Equal(ExtNamePinPolicy))
		assert.False(t, ExtNameTouchPolicy.Equal(ExtNamePinPolicy))
	})

	t.Run("OIDs share common prefix", func(t *testing.T) {
		// All share the prefix 1.3.6.1.4.1.65535.10
		assert.Equal(t, ExtNameTokenID[:len(ExtNameTokenID)-1], ExtNameTouchPolicy[:len(ExtNameTouchPolicy)-1])
		assert.Equal(t, ExtNameTokenID[:len(ExtNameTokenID)-1], ExtNamePinPolicy[:len(ExtNamePinPolicy)-1])
	})

	t.Run("OID final components", func(t *testing.T) {
		assert.Equal(t, 0, ExtNameTokenID[len(ExtNameTokenID)-1])
		assert.Equal(t, 1, ExtNameTouchPolicy[len(ExtNameTouchPolicy)-1])
		assert.Equal(t, 2, ExtNamePinPolicy[len(ExtNamePinPolicy)-1])
	})
}
