package yubikey

import (
	"crypto"
	"errors"
	"fmt"
	"log"

	"github.com/go-piv/piv-go/v2/piv"
)

type Yubikey struct {
	yk     *piv.YubiKey
	Serial uint32
}

func OpenBySerial(serial uint32) (*Yubikey, error) {
	cards, err := Cards()
	if err != nil {
		return nil, fmt.Errorf("failed to list cards: %w", err)
	}

	var card Card
	for _, row := range cards {
		if row.Serial == serial {
			card = row
			break
		}
	}

	if card.Serial == 0 {
		return nil, fmt.Errorf("yubikey with serial %d not found", serial)
	}

	return Open(card)
}

func Open(card Card) (*Yubikey, error) {
	yk, err := piv.Open(card.Name)
	if err != nil {
		return nil, err
	}

	version := yk.Version()
	if version.Major != 5 {
		return nil, fmt.Errorf("supported only Yubikey 5 version, current version: %d", version.Major)
	}

	if card.Serial != 0 {
		serial, err := yk.Serial()
		if err != nil {
			yk.Close()
			return nil, err
		}

		if serial != card.Serial {
			yk.Close()
			return nil, fmt.Errorf("serial number mismatch %d, got %d", card.Serial, serial)
		}
	}

	return &Yubikey{
		yk:     yk,
		Serial: card.Serial,
	}, nil
}

func (y *Yubikey) Close() error {
	if y.yk != nil {
		return y.yk.Close()
	}

	return nil
}

func (y *Yubikey) reOpen() error {
	if y.yk != nil {
		if _, err := y.yk.Serial(); err == nil {
			return nil
		}
	}

	if y.yk != nil {
		y.yk.Close()
		y.yk = nil
	}

	yk, err := OpenBySerial(y.Serial)
	if err != nil {
		return err
	}

	if y.Serial != yk.Serial {
		return fmt.Errorf("serial number mismatch %d, got %d", y.Serial, yk.Serial)
	}

	y.yk = yk.yk

	return nil
}

func (y *Yubikey) ResetToDefault() error {
	if err := y.reOpen(); err != nil {
		return err
	}

	if err := y.yk.Reset(); err != nil {
		return err
	}

	if err := y.yk.SetPIN(piv.DefaultPIN, piv.DefaultPIN); err != nil {
		return err
	}

	if err := y.yk.SetPUK(piv.DefaultPUK, piv.DefaultPUK); err != nil {
		return err
	}

	if err := y.yk.SetManagementKey(piv.DefaultManagementKey, piv.DefaultManagementKey); err != nil {
		return err
	}

	if err := y.yk.SetMetadata(piv.DefaultManagementKey, &piv.Metadata{
		ManagementKey: &piv.DefaultManagementKey,
	}); err != nil {
		return err
	}

	return nil
}

func (y *Yubikey) Reset(newPIN, newPUK string) error {
	if err := y.reOpen(); err != nil {
		return err
	}

	if err := y.yk.Reset(); err != nil {
		return err
	}

	if err := y.yk.SetPIN(piv.DefaultPIN, newPIN); err != nil {
		return err
	}

	if err := y.yk.SetPUK(piv.DefaultPUK, newPUK); err != nil {
		return err
	}

	if err := y.yk.VerifyPIN(newPIN); err != nil {
		return fmt.Errorf("failed to verify PIN: %w", err)
	}

	return nil
}

func (y *Yubikey) ResetMngmtKey(newKey []byte) error {
	if err := y.reOpen(); err != nil {
		return err
	}

	if err := y.yk.SetManagementKey(piv.DefaultManagementKey, newKey); err != nil {
		return err
	}

	meta := &piv.Metadata{
		ManagementKey: &newKey,
	}

	if err := y.yk.SetMetadata(newKey, meta); err != nil {
		return err
	}

	return nil
}

func (y *Yubikey) ListKeys(slots ...Slot) ([]Cert, error) {
	if err := y.reOpen(); err != nil {
		return nil, err
	}

	if len(slots) == 0 {
		slots = AllSlots
	}

	out := make([]Cert, 0, len(slots))

	for _, slot := range slots {
		cert, err := y.yk.Certificate(slot.PIVSlot)
		if err != nil {
			if !errors.Is(err, piv.ErrNotFound) {
				log.Printf("failed to get certificate from slot %d: %v\n", slot.PIVSlot.Key, err)
			}

			continue
		}

		out = append(out, Cert{
			Certificate: cert,
			Slot:        slot,
		})
	}

	return out, nil
}

func (y *Yubikey) getManagementKey(pin string) ([]byte, error) {
	if err := y.yk.VerifyPIN(pin); err != nil {
		return nil, fmt.Errorf("failed to verify PIN: %w", err)
	}

	meta, err := y.yk.Metadata(pin)
	if err != nil {
		return nil, err
	}

	if meta.ManagementKey == nil {
		return nil, errors.New("management key not set")
	}

	return *meta.ManagementKey, nil
}

func (y *Yubikey) PrivateKey(slot piv.Slot, public crypto.PublicKey, auth piv.KeyAuth) (crypto.PrivateKey, error) {
	if err := y.reOpen(); err != nil {
		return nil, err
	}

	return y.yk.PrivateKey(slot, public, auth)
}

func (y *Yubikey) Retries() (int, error) {
	if err := y.reOpen(); err != nil {
		return 0, err
	}

	return y.yk.Retries()
}

func (y *Yubikey) VerifyPIN(pin string) error {
	if err := y.reOpen(); err != nil {
		return err
	}

	return y.yk.VerifyPIN(pin)
}

func (y *Yubikey) SetPIN(currentPIN, newPIN string) error {
	if err := y.reOpen(); err != nil {
		return err
	}

	if _, err := y.yk.Retries(); err != nil {
		return err
	}

	return y.yk.SetPIN(currentPIN, newPIN)
}

func (y *Yubikey) SetPUK(currentPUK, newPUK string) error {
	if err := y.reOpen(); err != nil {
		return err
	}

	return y.yk.SetPUK(currentPUK, newPUK)
}

func (y *Yubikey) Unblock(pukCode, newPIN string) error {
	if err := y.reOpen(); err != nil {
		return err
	}

	return y.yk.Unblock(pukCode, newPIN)
}
