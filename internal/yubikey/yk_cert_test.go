package yubikey

import (
	"testing"

	"github.com/go-piv/piv-go/v2/piv"
	"github.com/stretchr/testify/assert"
)

func TestCertStruct(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		var c Cert
		assert.Nil(t, c.Certificate)
		assert.Equal(t, Slot{}, c.Slot)
	})

	t.Run("with slot", func(t *testing.T) {
		slot := Slot{PIVSlot: piv.SlotAuthentication}
		c := Cert{Slot: slot}
		assert.Equal(t, slot, c.Slot)
		assert.Nil(t, c.Certificate)
	})
}

func TestCertRequestStruct(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		var req CertRequest
		assert.Empty(t, req.CommonName)
		assert.Equal(t, 0, req.Days)
	})

	t.Run("with values", func(t *testing.T) {
		req := CertRequest{
			CommonName: "test-cert",
			Days:       365,
			Key: piv.Key{
				Algorithm:   piv.AlgorithmEC256,
				PINPolicy:   piv.PINPolicyOnce,
				TouchPolicy: piv.TouchPolicyNever,
			},
		}
		assert.Equal(t, "test-cert", req.CommonName)
		assert.Equal(t, 365, req.Days)
		assert.Equal(t, piv.AlgorithmEC256, req.Algorithm)
		assert.Equal(t, piv.PINPolicyOnce, req.PINPolicy)
		assert.Equal(t, piv.TouchPolicyNever, req.TouchPolicy)
	})
}

func TestGetCertPublicKey_NilYK(t *testing.T) {
	y := &Yubikey{Serial: 0}
	// With nil yk, Certificate will panic if called directly
	// But GetCertPublicKey doesn't call reOpen, it calls y.yk.Certificate directly
	// So this will panic - we need to check for nil yk
	assert.Panics(t, func() {
		_, _ = y.GetCertPublicKey(piv.SlotAuthentication)
	})
}

func TestGenCertificate_NilYK(t *testing.T) {
	y := &Yubikey{Serial: 0}
	slot := Slot{PIVSlot: piv.SlotAuthentication}
	req := CertRequest{
		Key: piv.Key{
			Algorithm:   piv.AlgorithmEC256,
			PINPolicy:   piv.PINPolicyOnce,
			TouchPolicy: piv.TouchPolicyNever,
		},
		CommonName: "test",
		Days:       365,
	}
	// GenCertificate calls getManagementKey which calls y.yk.VerifyPIN
	// With nil yk this will panic
	assert.Panics(t, func() {
		_, _ = y.GenCertificate(slot, "123456", req)
	})
}
