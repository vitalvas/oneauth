package yubikey

import "regexp"

var (
	validatePinRegex = regexp.MustCompile("^[0-9]{6,8}$")
)

func ValidatePin(pin string) bool {
	return validatePinRegex.MatchString(pin)
}

func ValidatePuk(puk string) bool {
	return validatePinRegex.MatchString(puk)
}
