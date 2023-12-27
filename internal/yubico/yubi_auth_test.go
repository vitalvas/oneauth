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
