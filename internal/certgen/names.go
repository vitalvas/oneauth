package certgen

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
)

var (
	ExtNameTokenID     = asn1.ObjectIdentifier([]int{1, 3, 6, 1, 4, 1, 65535, 10, 0})
	ExtNameTouchPolicy = asn1.ObjectIdentifier([]int{1, 3, 6, 1, 4, 1, 65535, 10, 1})
	ExtNamePinPolicy   = asn1.ObjectIdentifier([]int{1, 3, 6, 1, 4, 1, 65535, 10, 2})
)

type ExtraName struct {
	TokenID     string
	TouchPolicy string
	PinPolicy   string
}

func ParseExtraNames(names []pkix.AttributeTypeAndValue) (*ExtraName, error) {
	var out ExtraName

	for _, name := range names {
		switch {
		case name.Type.Equal(ExtNameTokenID):
			v, ok := name.Value.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected value type for token ID: %T", name.Value)
			}

			out.TokenID = v

		case name.Type.Equal(ExtNameTouchPolicy):
			v, ok := name.Value.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected value type for touch policy: %T", name.Value)
			}

			out.TouchPolicy = v

		case name.Type.Equal(ExtNamePinPolicy):
			v, ok := name.Value.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected value type for pin policy: %T", name.Value)
			}

			out.PinPolicy = v
		}
	}

	return &out, nil
}
