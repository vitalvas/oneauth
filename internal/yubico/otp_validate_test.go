package yubico

import (
	"testing"
)

func TestValidateOTP(t *testing.T) {
	otp := map[string]int64{
		// owned keys
		"cccccbhuinjdhkhclghtejrntflfbevvrvfvtkffghkj": 24017794, // 5c nano
		"cccccbhuinjdtgrgjfnelevvfjhteujcdigiicvujvcl": 24017794, // 5c nano

		"ccccccnlrctbvtjgucihjthunectghivervfrnnikvtr": 12239057, // 5 nfc
		"ccccccnlrctbfindlgneifuiteefgggtltlufeccrujt": 12239057, // 5 nfc

		"cccccbjudbfivcjjnrghdhetftgdrnkgeikhcfurrcdv": 26091847, // 5c
		"cccccbjudbfigtctgrheiiivvdgutieecvhtbunuvhhr": 26091847, // 5c

		// unowned keys
		"ccccccccltncdjjifceergtnukivgiujhgehgnkrfcef": 44464,
		"ccccccbchvthlivuitriujjifivbvtrjkjfirllluurj": 1077206,
	}

	for otp, expected := range otp {
		result, err := ValidateOTP(otp)
		if err != nil {
			t.Fatal(err)
		}

		if result != expected {
			t.Fatalf("expected %d, got %d", expected, result)
		}
	}
}

func TestValidateOTPFail(t *testing.T) {
	otp := map[string]error{
		"qwerty": ErrOTPHasInvalidLength,
		"cccccbhuinjdhkhclghtejrntflfbevvrvfvtkffghzz": ErrWrongOTPFormat,
	}

	for otp, expected := range otp {
		_, err := ValidateOTP(otp)
		if err != expected {
			t.Fatalf("expected %s, got %s", expected, err)
		}
	}
}
