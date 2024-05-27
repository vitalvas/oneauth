package tools

import (
	"encoding/hex"
	"testing"
)

func TestEncodePassphrase(t *testing.T) {
	testCases := []struct {
		input    []byte
		expected string
	}{
		{[]byte(""), "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{[]byte("hello"), "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{[]byte("world"), "486ea46224d1bb4fb680f34f7c9ad96a8f24ec88be73ea8e5a6c65260e9cb8a7"},
	}

	for _, tc := range testCases {
		result := EncodePassphrase(tc.input)
		resultEncoded := hex.EncodeToString(result)

		if resultEncoded != tc.expected {
			t.Errorf("EncodePassphrase(%s) = %s, expected %s", tc.input, resultEncoded, tc.expected)
		}
	}
}
