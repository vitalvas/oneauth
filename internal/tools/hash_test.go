package tools

import (
	"fmt"
	"testing"
)

func TestFastHash(t *testing.T) {
	tests := []struct {
		input    []byte
		expected string
	}{
		{[]byte(""), "cbf29ce484222325"},
		{[]byte("hello"), "a430d84680aabd0b"},
		{[]byte("test"), "f9e6e6ef197c2b25"},
		{[]byte("hello world"), "779a65e7023cd2e7"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Input: %s", test.input), func(t *testing.T) {
			result, err := FastHash(test.input)
			if err != nil {
				t.Fatalf("Error occurred: %v", err)
			}

			if result != test.expected {
				t.Errorf("Expected: %s, Got: %s", test.expected, result)
			}
		})
	}
}
