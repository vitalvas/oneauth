package yubikey

import (
	"fmt"

	"github.com/go-piv/piv-go/piv"
)

var (
	SlotKeyRSAID   = uint32(0x95)
	SlotKeyECDSAID = uint32(0x94)

	SlotKeyRSA   = MustSlotFromKeyID(SlotKeyRSAID)
	SlotKeyECDSA = MustSlotFromKeyID(SlotKeyECDSAID)

	AllSSHSlots = []Slot{
		SlotKeyRSA,
		SlotKeyECDSA,
	}

	AllSlots = func() []Slot {
		out := []Slot{
			MustSlotFromKeyID(piv.SlotAuthentication.Key),
			MustSlotFromKeyID(piv.SlotSignature.Key),
			MustSlotFromKeyID(piv.SlotKeyManagement.Key),
			MustSlotFromKeyID(piv.SlotCardAuthentication.Key),
		}

		for id := uint32(0x82); id <= 0x95; id++ {
			out = append(out, MustSlotFromKeyID(id))
		}

		return out
	}()
)

type Slot struct {
	PIVSlot piv.Slot
}

func (s Slot) String() string {
	return s.PIVSlot.String()
}

func MustSlotFromKeyID(keyID uint32) Slot {
	slot, err := SlotFromKeyID(keyID)
	if err != nil {
		panic(err)
	}

	return slot
}

func SlotFromKeyID(keyID uint32) (Slot, error) {
	switch keyID {
	case piv.SlotAuthentication.Key:
		return Slot{PIVSlot: piv.SlotAuthentication}, nil

	case piv.SlotSignature.Key:
		return Slot{PIVSlot: piv.SlotSignature}, nil

	case piv.SlotKeyManagement.Key:
		return Slot{PIVSlot: piv.SlotKeyManagement}, nil

	case piv.SlotCardAuthentication.Key:
		return Slot{PIVSlot: piv.SlotCardAuthentication}, nil
	}

	pivSlot, ok := piv.RetiredKeyManagementSlot(keyID)
	if !ok {
		return Slot{}, fmt.Errorf("unsupported key ID: 0x%02x", keyID)
	}

	return Slot{PIVSlot: pivSlot}, nil
}
