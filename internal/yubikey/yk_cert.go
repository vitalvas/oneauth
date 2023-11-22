package yubikey

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	"github.com/go-piv/piv-go/piv"
	"github.com/vitalvas/oneauth/internal/certgen"
)

func (y *Yubikey) GetCertPublicKey(slot piv.Slot) (crypto.PublicKey, error) {
	cert, err := y.yk.Certificate(slot)
	if err != nil {
		return nil, err
	}

	switch cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return cert.PublicKey, nil

	case *ecdsa.PublicKey:
		return cert.PublicKey, nil

	default:
		return nil, fmt.Errorf("unexpected public key type: %T", cert.PublicKey)
	}
}

func (y *Yubikey) GenCertificate(slot Slot, pin string, req CertRequest) (*x509.Certificate, error) {
	mgmtKey, err := y.getManagementKey(pin)
	if err != nil {
		return nil, err
	}

	pub, err := y.yk.GenerateKey(mgmtKey, slot.PIVSlot, req.Key)
	if err != nil {
		return nil, err
	}

	return y.setupCertificate(mgmtKey, slot, req.CommonName, pub, req.Days)
}

func (y *Yubikey) setupCertificate(mgmtKey [24]byte, slot Slot, commonName string, pub crypto.PublicKey, validDays int) (*x509.Certificate, error) {
	certBytes, err := certgen.GenCertificateFor(commonName, pub, validDays)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, err
	}

	if err := y.yk.SetCertificate(mgmtKey, slot.PIVSlot, cert); err != nil {
		return nil, fmt.Errorf("failed to set certificate: %w", err)
	}

	return cert, nil
}
