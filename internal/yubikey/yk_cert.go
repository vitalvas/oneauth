package yubikey

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
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

	var touchPolicy string
	switch req.TouchPolicy {
	case piv.TouchPolicyNever:
		touchPolicy = "never"

	case piv.TouchPolicyAlways:
		touchPolicy = "always"

	case piv.TouchPolicyCached:
		touchPolicy = "cached"

	default:
		touchPolicy = "-"
	}

	extraNames := []pkix.AttributeTypeAndValue{
		{
			Type:  certgen.ExtNameTokenID,
			Value: fmt.Sprintf("yubikey-%d", y.Serial),
		},
		{
			Type:  certgen.ExtNameTouchPolicy,
			Value: touchPolicy,
		},
	}

	certBytes, err := certgen.GenCertificateFor(req.CommonName, pub, req.Days, extraNames)
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
