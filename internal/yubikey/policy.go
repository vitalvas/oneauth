package yubikey

import (
	"github.com/go-piv/piv-go/piv"
)

var (
	toPINPolicy = map[string]piv.PINPolicy{
		"never":  piv.PINPolicyNever,
		"once":   piv.PINPolicyOnce,
		"always": piv.PINPolicyAlways,
	}
	toTouchPolicy = map[string]piv.TouchPolicy{
		"never":  piv.TouchPolicyNever,
		"always": piv.TouchPolicyAlways,
		"cached": piv.TouchPolicyCached,
	}
)

func MapPINPolicy(name string) (piv.PINPolicy, bool) {
	policy, ok := toPINPolicy[name]
	return policy, ok
}

func MapToStrPINPolicy(policy piv.PINPolicy) (string, bool) {
	for k, v := range toPINPolicy {
		if v == policy {
			return k, true
		}
	}

	return "-", false
}

func MapTouchPolicy(name string) (piv.TouchPolicy, bool) {
	policy, ok := toTouchPolicy[name]
	return policy, ok
}

func MapToStrTouchPolicy(policy piv.TouchPolicy) (string, bool) {
	for k, v := range toTouchPolicy {
		if v == policy {
			return k, true
		}
	}

	return "-", false
}
