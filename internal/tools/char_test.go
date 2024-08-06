package tools

import "testing"

func TestCountUniqueChars(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"abcdef", 6},
		{"aaaaa", 1},
		{"", 0},
		{"123123123", 3},
		{"11111111", 1},
		{"1212121212", 2},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := CountUniqueChars(test.input)
			if result != test.expected {
				t.Errorf("For input '%s', expected %d but got %d", test.input, test.expected, result)
			}
		})
	}
}
