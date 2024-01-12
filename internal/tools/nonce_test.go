package tools

import (
	"fmt"
	"testing"
)

func TestGenerateNonce(t *testing.T) {
	for _, test := range []int{0, 1, 10, 16, 32} {
		t.Run(fmt.Sprintf("Length %d", test), func(t *testing.T) {
			nonce, err := GenerateNonce(test)

			if err != nil {
				t.Errorf("GenerateNonce(%d) returned an error: %v", test, err)
			}

			if len(nonce) != test {
				t.Errorf("GenerateNonce(%d) returned a nonce of length %d, expected length %d", test, len(nonce), test)
			}

			for _, char := range nonce {
				if !isValidChar(string(char)) {
					t.Errorf("GenerateNonce(%d) returned a nonce containing an invalid character: %c", test, char)
				}
			}
		})
	}
}

func isValidChar(char string) bool {
	for _, allowedChar := range nonceAllowedChars {
		if string(allowedChar) == char {
			return true
		}
	}
	return false
}
