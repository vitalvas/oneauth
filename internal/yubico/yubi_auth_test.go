package yubico

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
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
