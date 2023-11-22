package yubikey

import (
	"crypto/x509"

	"github.com/go-piv/piv-go/piv"
)

type Cert struct {
	*x509.Certificate
	Slot Slot
}

type CertRequest struct {
	piv.Key
	CommonName string
	Days       int
}
