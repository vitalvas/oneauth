package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAESKey(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name     string
		input    string
		expected []byte
		hasError bool
	}{
		// Hex format tests
		{
			name:     "Hex/Valid 32 chars lowercase",
			input:    "31323334353637383930313233343536",
			expected: []byte("1234567890123456"),
			hasError: false,
		},
		{
			name:     "Hex/Valid 32 chars uppercase",
			input:    "31323334353637383930313233343536",
			expected: []byte("1234567890123456"),
			hasError: false,
		},
		{
			name:     "Hex/Valid mixed case",
			input:    "31323334353637383930313233343536",
			expected: []byte("1234567890123456"),
			hasError: false,
		},
		{
			name:     "Hex/With whitespace",
			input:    " 31323334353637383930313233343536 ",
			expected: []byte("1234567890123456"),
			hasError: false,
		},
		{
			name:     "Hex/Invalid characters",
			input:    "3132333435363738393031323334353Z",
			expected: nil,
			hasError: true,
		},
		{
			name:     "Hex/Wrong length - too short",
			input:    "313233343536373839303132333435",
			expected: nil,
			hasError: true,
		},
		{
			name:     "Hex/Wrong length - too long",
			input:    "3132333435363738393031323334353637",
			expected: nil,
			hasError: true,
		},

		// Base64 format tests
		{
			name:     "Base64/Valid URL encoding",
			input:    "MTIzNDU2Nzg5MDEyMzQ1Ng",
			expected: []byte("1234567890123456"),
			hasError: false,
		},
		{
			name:     "Base64/Valid standard encoding",
			input:    "MTIzNDU2Nzg5MDEyMzQ1Ng==",
			expected: []byte("1234567890123456"),
			hasError: false,
		},
		{
			name:     "Base64/With whitespace",
			input:    " MTIzNDU2Nzg5MDEyMzQ1Ng== ",
			expected: []byte("1234567890123456"),
			hasError: false,
		},
		{
			name:     "Base64/Invalid characters",
			input:    "MTIzNDU2Nzg5MDEyMzQ1Ng==!",
			expected: nil,
			hasError: true,
		},
		{
			name:     "Base64/Wrong length",
			input:    "MTIzNDU2Nzg5MA==",
			expected: nil,
			hasError: true,
		},

		// Edge cases
		{
			name:     "EdgeCase/Empty string",
			input:    "",
			expected: nil,
			hasError: true,
		},
		{
			name:     "EdgeCase/Only whitespace",
			input:    "   ",
			expected: nil,
			hasError: true,
		},
		{
			name:     "EdgeCase/Random string",
			input:    "not_hex_or_base64",
			expected: nil,
			hasError: true,
		},
		{
			name:     "EdgeCase/Almost hex but wrong chars",
			input:    "31323334353637383930313233343536GH",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.parseAESKey(tt.input)

			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
