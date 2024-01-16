package yubikey

import (
	"fmt"
	"testing"

	"github.com/go-piv/piv-go/piv"
)

func TestSlotFromKeyID(t *testing.T) {
	tests := []struct {
		name   string
		keyID  uint32
		result Slot
		errMsg string
	}{
		{
			name:   "Valid Authentication Key ID",
			keyID:  piv.SlotAuthentication.Key,
			result: Slot{PIVSlot: piv.SlotAuthentication},
		},
		{
			name:   "Valid Signature Key ID",
			keyID:  piv.SlotSignature.Key,
			result: Slot{PIVSlot: piv.SlotSignature},
		},
		{
			name:   "Valid Key Management Key ID",
			keyID:  piv.SlotKeyManagement.Key,
			result: Slot{PIVSlot: piv.SlotKeyManagement},
		},
		{
			name:   "Valid Card Authentication Key ID",
			keyID:  piv.SlotCardAuthentication.Key,
			result: Slot{PIVSlot: piv.SlotCardAuthentication},
		},
		{
			name:   "Unsupported Key ID",
			keyID:  0x10, // Replace with an unsupported key ID
			result: Slot{},
			errMsg: "unsupported key ID: 0x10",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			slot, err := SlotFromKeyID(test.keyID)

			if fmt.Sprintf("%v", slot) != fmt.Sprintf("%v", test.result) {
				t.Errorf("Expected result: %v, but got: %v", test.result, slot)
			}

			if err == nil && test.errMsg != "" {
				t.Errorf("Expected error message: %v, but got no error", test.errMsg)
			}

			if err != nil && err.Error() != test.errMsg {
				t.Errorf("Expected error message: %v, but got: %v", test.errMsg, err.Error())
			}
		})
	}

	for id := uint32(0x82); id <= 0x95; id++ {
		t.Run(fmt.Sprintf("Valid Retired Key Management Key ID 0x%02x", id), func(t *testing.T) {
			slot, err := SlotFromKeyID(id)

			if err != nil {
				t.Errorf("Expected no error, but got: %v", err.Error())
			}

			if slot.PIVSlot.Key != id {
				t.Errorf("Expected result: %v, but got: %v", id, slot.PIVSlot.Key)
			}
		})
	}

	for _, id := range []uint32{
		0xf9, // attestation
		0xff, // undefined
	} {
		t.Run(fmt.Sprintf("Valid Retired Key Management Key ID 0x%02x", id), func(t *testing.T) {
			_, err := SlotFromKeyID(id)

			if err == nil {
				t.Errorf("Expected error, but got none")
			}
		})
	}
}
