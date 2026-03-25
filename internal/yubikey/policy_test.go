package yubikey

import (
	"testing"

	"github.com/go-piv/piv-go/v2/piv"
)

func TestMapPINPolicy(t *testing.T) {
	testCases := []struct {
		name        string
		expected    piv.PINPolicy
		shouldExist bool
	}{
		{"never", piv.PINPolicyNever, true},
		{"once", piv.PINPolicyOnce, true},
		{"always", piv.PINPolicyAlways, true},
		{"invalid", 0, false}, // Test for a non-existent policy
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policy, exists := MapPINPolicy(tc.name)

			if exists != tc.shouldExist {
				t.Errorf("Expected existence to be %v, but got %v", tc.shouldExist, exists)
			}

			if policy != tc.expected {
				t.Errorf("Expected policy to be %v, but got %v", tc.expected, policy)
			}
		})
	}
}

func TestMapToStrPINPolicy(t *testing.T) {
	testCases := []struct {
		policy       piv.PINPolicy
		expectedName string
		shouldExist  bool
	}{
		{piv.PINPolicyNever, "never", true},
		{piv.PINPolicyOnce, "once", true},
		{piv.PINPolicyAlways, "always", true},
		{piv.PINPolicy(42), "-", false}, // Test for a non-existent policy
	}

	for _, tc := range testCases {
		t.Run(tc.expectedName, func(t *testing.T) {
			name, exists := MapToStrPINPolicy(tc.policy)

			if exists != tc.shouldExist {
				t.Errorf("Expected existence to be %v, but got %v", tc.shouldExist, exists)
			}

			if name != tc.expectedName {
				t.Errorf("Expected name to be %s, but got %s", tc.expectedName, name)
			}
		})
	}
}

func TestMapTouchPolicy(t *testing.T) {
	testCases := []struct {
		name        string
		expected    piv.TouchPolicy
		shouldExist bool
	}{
		{"never", piv.TouchPolicyNever, true},
		{"always", piv.TouchPolicyAlways, true},
		{"cached", piv.TouchPolicyCached, true},
		{"nonexistent", 0, false}, // Test for a non-existent policy
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			policy, exists := MapTouchPolicy(testCase.name)

			if exists != testCase.shouldExist {
				t.Errorf("Expected existence: %v, Got: %v", testCase.shouldExist, exists)
			}

			if policy != testCase.expected {
				t.Errorf("Expected policy: %v, Got: %v", testCase.expected, policy)
			}
		})
	}
}

func TestMapToStrTouchPolicy(t *testing.T) {
	testCases := []struct {
		input    piv.TouchPolicy
		expected string
		exists   bool
	}{
		{piv.TouchPolicyNever, "never", true},
		{piv.TouchPolicyAlways, "always", true},
		{piv.TouchPolicyCached, "cached", true},
		{piv.TouchPolicy(100), "-", false}, // Test for a non-existent policy
	}

	for _, testCase := range testCases {
		t.Run(testCase.expected, func(t *testing.T) {
			str, exists := MapToStrTouchPolicy(testCase.input)

			if exists != testCase.exists {
				t.Errorf("Expected existence: %v, Got: %v", testCase.exists, exists)
			}

			if str != testCase.expected {
				t.Errorf("Expected string: %v, Got: %v", testCase.expected, str)
			}
		})
	}
}
