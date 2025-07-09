package yubikey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/go-piv/piv-go/piv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCert_Structure(t *testing.T) {
	// Create a test certificate
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test-cert",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	x509Cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	slot := Slot{
		PIVSlot: piv.SlotAuthentication,
	}

	cert := Cert{
		Certificate: x509Cert,
		Slot:        slot,
	}

	// Test that the embedded certificate works
	assert.Equal(t, "test-cert", cert.Subject.CommonName)
	assert.Equal(t, slot, cert.Slot)
	assert.Equal(t, x509Cert, cert.Certificate)
}

func TestCertRequest_Structure(t *testing.T) {
	key := piv.Key{
		Algorithm:   piv.AlgorithmRSA2048,
		PINPolicy:   piv.PINPolicyAlways,
		TouchPolicy: piv.TouchPolicyNever,
	}
	
	certReq := CertRequest{
		Key:        key,
		CommonName: "test@example.com",
		Days:       365,
	}

	assert.Equal(t, key, certReq.Key)
	assert.Equal(t, "test@example.com", certReq.CommonName)
	assert.Equal(t, 365, certReq.Days)
}

func TestCertRequest_DefaultValues(t *testing.T) {
	var certReq CertRequest

	assert.Equal(t, piv.Key{}, certReq.Key)
	assert.Equal(t, "", certReq.CommonName)
	assert.Equal(t, 0, certReq.Days)
}

func TestCertRequest_WithKeyTypes(t *testing.T) {
	tests := []struct {
		name    string
		keyType piv.Key
	}{
		{
			name: "RSA2048",
			keyType: piv.Key{
				Algorithm:   piv.AlgorithmRSA2048,
				PINPolicy:   piv.PINPolicyAlways,
				TouchPolicy: piv.TouchPolicyNever,
			},
		},
		{
			name: "ECCP256",
			keyType: piv.Key{
				Algorithm:   piv.AlgorithmEC256,
				PINPolicy:   piv.PINPolicyAlways,
				TouchPolicy: piv.TouchPolicyNever,
			},
		},
		{
			name: "ECCP384",
			keyType: piv.Key{
				Algorithm:   piv.AlgorithmEC384,
				PINPolicy:   piv.PINPolicyAlways,
				TouchPolicy: piv.TouchPolicyNever,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			certReq := CertRequest{
				Key:        tt.keyType,
				CommonName: "test-user",
				Days:       30,
			}

			assert.Equal(t, tt.keyType, certReq.Key)
			assert.Equal(t, "test-user", certReq.CommonName)
			assert.Equal(t, 30, certReq.Days)
		})
	}
}

func TestCert_WithDifferentSlots(t *testing.T) {
	// Create a simple test certificate
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour),
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	x509Cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	tests := []struct {
		name string
		slot Slot
	}{
		{
			name: "Authentication slot",
			slot: Slot{
				PIVSlot: piv.SlotAuthentication,
			},
		},
		{
			name: "Signature slot",
			slot: Slot{
				PIVSlot: piv.SlotSignature,
			},
		},
		{
			name: "Key Management slot",
			slot: Slot{
				PIVSlot: piv.SlotKeyManagement,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert := Cert{
				Certificate: x509Cert,
				Slot:        tt.slot,
			}

			assert.Equal(t, x509Cert, cert.Certificate)
			assert.Equal(t, tt.slot, cert.Slot)
			assert.Equal(t, tt.slot.PIVSlot, cert.Slot.PIVSlot)
		})
	}
}

func TestCertRequest_ValidationFields(t *testing.T) {
	// Test with various common names and day values
	tests := []struct {
		name       string
		commonName string
		days       int
		valid      bool
	}{
		{
			name:       "valid email",
			commonName: "user@example.com",
			days:       365,
			valid:      true,
		},
		{
			name:       "valid hostname",
			commonName: "server.example.com",
			days:       90,
			valid:      true,
		},
		{
			name:       "empty common name",
			commonName: "",
			days:       30,
			valid:      false, // Usually invalid but depends on use case
		},
		{
			name:       "zero days",
			commonName: "test",
			days:       0,
			valid:      false, // Usually invalid
		},
		{
			name:       "negative days",
			commonName: "test",
			days:       -1,
			valid:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			certReq := CertRequest{
				Key: piv.Key{
					Algorithm:   piv.AlgorithmEC256,
					PINPolicy:   piv.PINPolicyAlways,
					TouchPolicy: piv.TouchPolicyNever,
				},
				CommonName: tt.commonName,
				Days:       tt.days,
			}

			// Basic validation - ensure structure is created correctly
			assert.Equal(t, tt.commonName, certReq.CommonName)
			assert.Equal(t, tt.days, certReq.Days)

			// Validate business rules
			if tt.valid {
				assert.NotEmpty(t, certReq.CommonName, "Valid requests should have non-empty common name")
				assert.Positive(t, certReq.Days, "Valid requests should have positive days")
			}
		})
	}
}