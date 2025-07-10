package certgen

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenCertificateFor_WithECDSAKey(t *testing.T) {
	// Generate test ECDSA key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	publicKey := privateKey.Public()
	commonName := "test-user@example.com"
	days := 365

	certBytes, err := GenCertificateFor(commonName, publicKey, days, nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, certBytes)

	// Parse and validate the certificate
	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	assert.Equal(t, commonName, cert.Subject.CommonName)
	assert.True(t, cert.NotBefore.Before(time.Now().Add(time.Minute)))
	assert.True(t, cert.NotAfter.After(time.Now().Add(time.Duration(days-1)*24*time.Hour)))
	assert.Contains(t, cert.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	assert.Equal(t, x509.KeyUsageDigitalSignature|x509.KeyUsageCertSign, cert.KeyUsage)
	assert.True(t, cert.BasicConstraintsValid)
}

func TestGenCertificateFor_WithRSAKey(t *testing.T) {
	// Generate test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicKey := privateKey.Public()
	commonName := "rsa-user@example.com"
	days := 180

	certBytes, err := GenCertificateFor(commonName, publicKey, days, nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, certBytes)

	// Parse and validate the certificate
	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	assert.Equal(t, commonName, cert.Subject.CommonName)
	assert.True(t, cert.NotBefore.Before(time.Now().Add(time.Minute)))
	assert.True(t, cert.NotAfter.After(time.Now().Add(time.Duration(days-1)*24*time.Hour)))
}

func TestGenCertificateFor_WithExtraNames(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	publicKey := privateKey.Public()
	commonName := "test-user"
	extraNames := []pkix.AttributeTypeAndValue{
		{
			Type:  []int{2, 5, 4, 6}, // Country
			Value: "US",
		},
		{
			Type:  []int{2, 5, 4, 10}, // Organization
			Value: "Test Org",
		},
	}

	certBytes, err := GenCertificateFor(commonName, publicKey, 365, extraNames)

	assert.NoError(t, err)
	assert.NotEmpty(t, certBytes)

	// Parse and validate the certificate
	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	assert.Equal(t, commonName, cert.Subject.CommonName)
	// Note: ExtraNames are passed to the certificate but may not be directly accessible in parsed certificate
}

func TestGenCertificateFor_DifferentDays(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	publicKey := privateKey.Public()
	commonName := "test-user"

	tests := []struct {
		name string
		days int
	}{
		{"1 day", 1},
		{"30 days", 30},
		{"365 days", 365},
		{"730 days", 730},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			certBytes, err := GenCertificateFor(commonName, publicKey, tt.days, nil)

			assert.NoError(t, err)
			assert.NotEmpty(t, certBytes)

			cert, err := x509.ParseCertificate(certBytes)
			require.NoError(t, err)

			expectedExpiry := time.Now().AddDate(0, 0, tt.days)
			timeDiff := cert.NotAfter.Sub(expectedExpiry)

			// Allow some tolerance for test execution time
			assert.True(t, timeDiff < time.Minute && timeDiff > -time.Minute,
				"Certificate expiry time should be close to expected: got %v, expected %v",
				cert.NotAfter, expectedExpiry)
		})
	}
}

func TestGenCertificateFor_SerialNumberUnique(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	publicKey := privateKey.Public()
	commonName := "test-user"

	// Generate multiple certificates
	var serialNumbers []string
	for i := 0; i < 5; i++ {
		certBytes, err := GenCertificateFor(commonName, publicKey, 365, nil)
		require.NoError(t, err)

		cert, err := x509.ParseCertificate(certBytes)
		require.NoError(t, err)

		serialNumbers = append(serialNumbers, cert.SerialNumber.String())

		// Small delay to ensure different timestamps
		time.Sleep(time.Millisecond)
	}

	// Check that all serial numbers are unique
	serialSet := make(map[string]bool)
	for _, serial := range serialNumbers {
		assert.False(t, serialSet[serial], "Serial number should be unique: %s", serial)
		serialSet[serial] = true
	}
}

func TestGenCertificateFor_CertificateStructure(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	publicKey := privateKey.Public()
	commonName := "structure-test"

	certBytes, err := GenCertificateFor(commonName, publicKey, 365, nil)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	// Verify certificate structure
	assert.Equal(t, commonName, cert.Subject.CommonName)
	assert.NotNil(t, cert.SerialNumber)
	assert.True(t, cert.SerialNumber.Sign() > 0)
	assert.True(t, cert.NotBefore.Before(cert.NotAfter))
	assert.Contains(t, cert.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	assert.Equal(t, x509.KeyUsageDigitalSignature|x509.KeyUsageCertSign, cert.KeyUsage)
	assert.True(t, cert.BasicConstraintsValid)

	// Verify the public key matches
	assert.Equal(t, publicKey, cert.PublicKey)
}

func TestGenCertificateFor_EmptyCommonName(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	publicKey := privateKey.Public()

	certBytes, err := GenCertificateFor("", publicKey, 365, nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, certBytes)

	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	assert.Empty(t, cert.Subject.CommonName)
}

func TestGenCertificateFor_ZeroDays(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	publicKey := privateKey.Public()

	certBytes, err := GenCertificateFor("test", publicKey, 0, nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, certBytes)

	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	// Certificate should be valid for exactly the same day (0 days)
	assert.True(t, cert.NotAfter.After(cert.NotBefore) || cert.NotAfter.Equal(cert.NotBefore))
	// When days=0, AddDate(0,0,0) gives same day but same time, so duration should be very small
	duration := cert.NotAfter.Sub(cert.NotBefore)
	assert.True(t, duration <= 24*time.Hour)
}

func TestGenCommonName(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		hostName string
		expected string
	}{
		{
			name:     "Basic format",
			user:     "testuser",
			hostName: "example.com",
			expected: "testuser@example.com",
		},
		{
			name:     "Empty user",
			user:     "",
			hostName: "example.com",
			expected: "@example.com",
		},
		{
			name:     "Empty hostname",
			user:     "testuser",
			hostName: "",
			expected: "testuser@",
		},
		{
			name:     "Both empty",
			user:     "",
			hostName: "",
			expected: "@",
		},
		{
			name:     "Special characters",
			user:     "test.user+tag",
			hostName: "sub.domain.example.com",
			expected: "test.user+tag@sub.domain.example.com",
		},
		{
			name:     "Numbers and hyphens",
			user:     "user123",
			hostName: "host-1.example-domain.org",
			expected: "user123@host-1.example-domain.org",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenCommonName(tt.user, tt.hostName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseExtraNames(t *testing.T) {
	tests := []struct {
		name        string
		extraNames  []pkix.AttributeTypeAndValue
		expected    *ExtraName
		expectError bool
	}{
		{
			name:       "empty-names",
			extraNames: []pkix.AttributeTypeAndValue{},
			expected: &ExtraName{
				TokenID:     "",
				TouchPolicy: "",
				PinPolicy:   "",
			},
			expectError: false,
		},
		{
			name: "token-id-only",
			extraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  ExtNameTokenID,
					Value: "12345",
				},
			},
			expected: &ExtraName{
				TokenID:     "12345",
				TouchPolicy: "",
				PinPolicy:   "",
			},
			expectError: false,
		},
		{
			name: "touch-policy-only",
			extraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  ExtNameTouchPolicy,
					Value: "always",
				},
			},
			expected: &ExtraName{
				TokenID:     "",
				TouchPolicy: "always",
				PinPolicy:   "",
			},
			expectError: false,
		},
		{
			name: "pin-policy-only",
			extraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  ExtNamePinPolicy,
					Value: "once",
				},
			},
			expected: &ExtraName{
				TokenID:     "",
				TouchPolicy: "",
				PinPolicy:   "once",
			},
			expectError: false,
		},
		{
			name: "all-fields",
			extraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  ExtNameTokenID,
					Value: "98765",
				},
				{
					Type:  ExtNameTouchPolicy,
					Value: "cached",
				},
				{
					Type:  ExtNamePinPolicy,
					Value: "never",
				},
			},
			expected: &ExtraName{
				TokenID:     "98765",
				TouchPolicy: "cached",
				PinPolicy:   "never",
			},
			expectError: false,
		},
		{
			name: "unknown-oid",
			extraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  []int{1, 2, 3, 4, 5}, // Unknown OID
					Value: "unknown",
				},
			},
			expected: &ExtraName{
				TokenID:     "",
				TouchPolicy: "",
				PinPolicy:   "",
			},
			expectError: false, // Unknown OIDs are ignored
		},
		{
			name: "invalid-token-id-type",
			extraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  ExtNameTokenID,
					Value: 12345, // Should be string, not int
				},
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "invalid-touch-policy-type",
			extraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  ExtNameTouchPolicy,
					Value: true, // Should be string, not bool
				},
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "invalid-pin-policy-type",
			extraNames: []pkix.AttributeTypeAndValue{
				{
					Type:  ExtNamePinPolicy,
					Value: []byte("bytes"), // Should be string, not []byte
				},
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseExtraNames(tt.extraNames)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestExtNameOIDs(t *testing.T) {
	// Test that the OID constants are properly defined
	assert.Equal(t, []int{1, 3, 6, 1, 4, 1, 65535, 10, 0}, []int(ExtNameTokenID))
	assert.Equal(t, []int{1, 3, 6, 1, 4, 1, 65535, 10, 1}, []int(ExtNameTouchPolicy))
	assert.Equal(t, []int{1, 3, 6, 1, 4, 1, 65535, 10, 2}, []int(ExtNamePinPolicy))
}

func TestExtraNameStruct(t *testing.T) {
	extraName := ExtraName{
		TokenID:     "123456",
		TouchPolicy: "always",
		PinPolicy:   "once",
	}

	assert.Equal(t, "123456", extraName.TokenID)
	assert.Equal(t, "always", extraName.TouchPolicy)
	assert.Equal(t, "once", extraName.PinPolicy)
}
