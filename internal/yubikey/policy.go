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
