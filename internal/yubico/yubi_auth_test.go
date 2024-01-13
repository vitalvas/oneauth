package yubico

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"testing"
)

func TestSignRequest(t *testing.T) {
	secret := []byte("secretKey")
	params := url.Values{}
	params.Set("param1", "value1")
	params.Set("param2", "value2")

	h := hmac.New(sha1.New, secret)
	u := params.Encode()
	h.Write([]byte(u))
	expectedSig := h.Sum(nil)
	expectedSigBase64 := base64.StdEncoding.EncodeToString(expectedSig)

	signRequest(params, secret)

	actualSigBase64 := params.Get("h")
	if actualSigBase64 != expectedSigBase64 {
		t.Errorf("Expected signature: %s, but got: %s", expectedSigBase64, actualSigBase64)
	}

	if params.Get("param1") != "value1" || params.Get("param2") != "value2" {
		t.Error("Other parameters have been modified")
	}
}

func TestSignEmptyRequest(t *testing.T) {
	secret := []byte("secretKey")
	params := url.Values{}

	signRequest(params, secret)

	actualSigBase64 := params.Get("h")
	if actualSigBase64 != "Uxv59Quy2jOAJnXqbfK9TYwfrvY=" {
		t.Errorf("Expected signature: Uxv59Quy2jOAJnXqbfK9TYwfrvY=, but got: %s", actualSigBase64)
	}
}

func TestResponseFromBody(t *testing.T) {
	input := []byte(`h=ZQTg6Vo/Ti7LFKi9x/K8te+9SKI=
t=2024-01-12T03:13:22Z0504
otp=cccccbhuinjdrvtgbgrbrcikvrtvulvltkdufcrngunn
nonce=askjdnkajsndjkasndkjsnad
sl=100
timestamp=4272362
sessioncounter=26
sessionuse=3
status=OK
wrong
`)

	resp, err := responseFromBody(input)

	if err != nil {
		t.Errorf("Expected no error, but got an error: %v", err)
	}

	expected := &VerifyResponse{
		Timestamp:      4272362,
		SessionCounter: 26,
		SessionUse:     3,
		Status:         "OK",
	}

	if resp.Timestamp != expected.Timestamp {
		t.Errorf("Expected Timestamp to be %d, but got %d", expected.Timestamp, resp.Timestamp)
	}

	if resp.SessionCounter != expected.SessionCounter {
		t.Errorf("Expected SessionCounter to be %d, but got %d", expected.SessionCounter, resp.SessionCounter)
	}

	if resp.SessionUse != expected.SessionUse {
		t.Errorf("Expected SessionUse to be %d, but got %d", expected.SessionUse, resp.SessionUse)
	}

	if resp.Status != expected.Status {
		t.Errorf("Expected Status to be %s, but got %s", expected.Status, resp.Status)
	}
}

func TestGetRequestParams(t *testing.T) {

	tests := []struct {
		clientID int
		otp      string
	}{
		{clientID: 123, otp: "cccccbhuinjdrvtgbgrbrcikvrtvulvltkdufcrngunn"},
		{clientID: 456, otp: "cccccbhuinjdrvtgbgrbrcikvrtvulvltkdufcrngunn"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("clientID=%d otp=%s", test.clientID, test.otp), func(t *testing.T) {
			params, err := getRequestParams(test.clientID, test.otp)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check individual values
			if gotID := params.Get("id"); gotID != fmt.Sprintf("%d", test.clientID) {
				t.Errorf("Expected id=%s, got id=%s", fmt.Sprintf("%d", test.clientID), gotID)
			}
			if gotOTP := params.Get("otp"); gotOTP != test.otp {
				t.Errorf("Expected otp=%s, got otp=%s", test.otp, gotOTP)
			}
			if gotNonce := params.Get("nonce"); len(gotNonce) != 32 {
				t.Errorf("Expected nonce size 32, got %d", len(gotNonce))
			}
			if gotTimestamp := params.Get("timestamp"); gotTimestamp != "1" {
				t.Errorf("Expected timestamp=1, got timestamp=%s", gotTimestamp)
			}
			if gotSL := params.Get("sl"); gotSL != "secure" {
				t.Errorf("Expected sl=secure, got sl=%s", gotSL)
			}
			if gotTimeout := params.Get("timeout"); gotTimeout != "2" {
				t.Errorf("Expected timeout=2, got timeout=%s", gotTimeout)
			}
		})
	}
}
