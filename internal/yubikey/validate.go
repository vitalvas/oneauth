package yubikey

import (
	"regexp"
	"slices"

	"github.com/vitalvas/oneauth/internal/tools"
)

var (
	validatePinRegex = regexp.MustCompile("^[0-9]{6,8}$")
	validatePukRegex = regexp.MustCompile("^[0-9]{8}$")
	pinBlocking      = []string{ // source: https://docs.yubico.com/hardware/yubikey/yk-tech-manual/5.7-firmware-specifics.html#pin-complexity
		"123456",
		"123123",
		"654321",
		"123321",
		"112233",
		"121212",
		"123456789",
		"12345678",
		"1234567",
		"520520",
		"123654",
		"1234567890",
		"159753",
	}
)

func ValidatePin(pin string) bool {
	if validatePinRegex.MatchString(pin) {
		if !slices.Contains(pinBlocking, pin) {
			return tools.CountUniqueChars(pin) > 1
		}
	}

	return false
}

func ValidatePuk(puk string) bool {
	if validatePukRegex.MatchString(puk) {
		if !slices.Contains(pinBlocking, puk) {
			return tools.CountUniqueChars(puk) > 1
		}
	}

	return false
}
