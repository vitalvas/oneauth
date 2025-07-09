package yubikey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCard_String(t *testing.T) {
	tests := []struct {
		name     string
		card     Card
		expected string
	}{
		{
			name: "Basic card",
			card: Card{
				Name:    "Yubico YubiKey FIDO+CCID 00 00",
				Serial:  12345678,
				Version: "5.4.3",
			},
			expected: "Yubikey #12345678",
		},
		{
			name: "Zero serial",
			card: Card{
				Name:    "test-card",
				Serial:  0,
				Version: "1.0.0",
			},
			expected: "Yubikey #0",
		},
		{
			name: "Large serial",
			card: Card{
				Name:    "another-card",
				Serial:  4294967295, // Max uint32
				Version: "6.0.0",
			},
			expected: "Yubikey #4294967295",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.card.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCard_Structure(t *testing.T) {
	card := Card{
		Name:    "Test YubiKey",
		Serial:  987654321,
		Version: "5.2.7",
	}

	assert.Equal(t, "Test YubiKey", card.Name)
	assert.Equal(t, uint32(987654321), card.Serial)
	assert.Equal(t, "5.2.7", card.Version)
}

func TestCards_NoYubikeys(t *testing.T) {
	// This test will likely fail in CI/testing environments where no YubiKeys are present
	// We mainly test that the function doesn't panic and returns appropriate results
	cards, err := Cards()

	// Either we get an error (no cards found, no readers, etc.) or empty slice
	if err != nil {
		t.Logf("Cards() returned error (expected in test environments): %v", err)
		assert.Nil(t, cards)
	} else {
		// If no error, cards should be a valid slice (possibly empty)
		assert.NotNil(t, cards)
		t.Logf("Found %d YubiKey cards", len(cards))
	}
}

func TestCardRead_InvalidName(t *testing.T) {
	// Test with a card name that doesn't exist
	card, err := cardRead("non-existent-card-name")

	assert.Error(t, err)
	assert.Nil(t, card)
	assert.Contains(t, err.Error(), "failed to open card")
}

func TestCardRead_EmptyName(t *testing.T) {
	// Test with empty card name
	card, err := cardRead("")

	assert.Error(t, err)
	assert.Nil(t, card)
}

// Mock tests to validate the logic without requiring actual hardware
func TestCard_StringConsistency(t *testing.T) {
	card := Card{
		Name:    "Mock YubiKey",
		Serial:  123456,
		Version: "5.4.3",
	}

	// Test that String() is consistent
	result1 := card.String()
	result2 := card.String()

	assert.Equal(t, result1, result2)
	assert.Equal(t, "Yubikey #123456", result1)
}

func TestCard_DefaultValues(t *testing.T) {
	// Test zero-value card
	var card Card

	assert.Equal(t, "", card.Name)
	assert.Equal(t, uint32(0), card.Serial)
	assert.Equal(t, "", card.Version)
	assert.Equal(t, "Yubikey #0", card.String())
}

func TestCard_FieldAssignment(t *testing.T) {
	card := Card{}

	// Test field assignment
	card.Name = "Assigned Name"
	card.Serial = 999888777
	card.Version = "1.2.3"

	assert.Equal(t, "Assigned Name", card.Name)
	assert.Equal(t, uint32(999888777), card.Serial)
	assert.Equal(t, "1.2.3", card.Version)
	assert.Equal(t, "Yubikey #999888777", card.String())
}
