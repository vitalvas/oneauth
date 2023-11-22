package certgen

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"
)

func GenCertificateFor(commonName string, pub crypto.PublicKey, days int) ([]byte, error) {
	var parentPriv crypto.PrivateKey
	var parentPub crypto.PublicKey

	switch pub.(type) {
	case *rsa.PublicKey:
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate private key: %w", err)
		}

		parentPub = priv.Public()
		parentPriv = priv

	case *ecdsa.PublicKey:
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate private key: %w", err)
		}

		parentPub = priv.Public()
		parentPriv = priv

	default:
		return nil, fmt.Errorf("unsupported public key type: %T", pub)
	}

	parent := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "OneAuth SSH Fake CA",
		},
		PublicKey: parentPub,
	}

	csr := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, days),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, csr, parent, pub, parentPriv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	return certBytes, nil
}
