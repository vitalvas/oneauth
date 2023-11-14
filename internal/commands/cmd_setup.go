package commands

import (
	"fmt"
	"time"

	"github.com/go-piv/piv-go/piv"
	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/certgen"
	"github.com/vitalvas/oneauth/internal/yubikey"
)

var setupCmd = &cli.Command{
	Name:  "setup",
	Usage: "Setup a new YubiKey",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:     "confirm",
			Usage:    "Confirm the setup (all data on the YubiKey will be wiped)",
			Required: true,
		},
		&cli.Uint64Flag{
			Name:  "wait",
			Value: 5,
		},
		&cli.Uint64Flag{
			Name:  "serial",
			Usage: "YubiKey serial number",
		},
		&cli.StringFlag{
			Name:  "username",
			Value: "oneauth",
		},
	},
	Before: selectYubiKey,
	Action: func(c *cli.Context) error {
		serial := c.Uint64("serial")
		if serial == 0 {
			return fmt.Errorf("serial is required")
		}

		key, err := yubikey.OpenBySerial(uint32(serial))
		if err != nil {
			return err
		}

		defer key.Close()

		fmt.Println("Serial:", serial)
		fmt.Println("Wipping the YubiKey...")

		if wait := c.Uint64("wait"); wait > 0 {
			fmt.Printf("Waiting %d seconds for cancel...\n", wait)
			time.Sleep(time.Duration(wait) * time.Second)
		}

		newPIN, err := yubikey.GeneratePinCode()
		if err != nil {
			return err
		}

		fmt.Println("[!] New YubiKey PIN:", newPIN)

		newPUK, err := yubikey.GeneratePukCode()
		if err != nil {
			return err
		}

		fmt.Println("[!] New YubiKey PUK:", newPUK)

		if err := key.Reset(newPIN, newPUK); err != nil {
			return err
		}

		newManagementKey, err := yubikey.GenerateManagementKey()
		if err != nil {
			return err
		}

		if err := key.ResetMngmtKey(newManagementKey); err != nil {
			return err
		}

		username := c.String("username")

		key.GenCertificate(yubikey.MustSlotFromKeyID(yubikey.SlotKeyRSAID), newPIN, yubikey.CertRequest{
			CommonName: certgen.GenCommonName(username, "insecure-rsa"),
			Key: piv.Key{
				Algorithm:   piv.AlgorithmRSA2048,
				PINPolicy:   piv.PINPolicyNever,
				TouchPolicy: piv.TouchPolicyNever,
			},
		})

		key.GenCertificate(yubikey.MustSlotFromKeyID(yubikey.SlotKeyECDSAID), newPIN, yubikey.CertRequest{
			CommonName: certgen.GenCommonName(username, "insecure-ecdsa"),
			Key: piv.Key{
				Algorithm:   piv.AlgorithmEC256,
				PINPolicy:   piv.PINPolicyNever,
				TouchPolicy: piv.TouchPolicyNever,
			},
		})

		return nil
	},
}

func selectYubiKey(c *cli.Context) error {
	cards, err := yubikey.Cards()
	if err != nil {
		return err
	}

	serial := c.Uint64("serial")
	if serial > 0 {
		for _, row := range cards {
			if row.Serial == uint32(serial) {
				if err := c.Set("serial", fmt.Sprintf("%d", serial)); err != nil {
					return err
				}
				return nil
			}
		}

		return fmt.Errorf("YubiKey with serial %d not found", serial)
	}

	if len(cards) != 1 {
		return fmt.Errorf("expected exactly one card, got %d", len(cards))
	}

	if err := c.Set("serial", fmt.Sprintf("%d", cards[0].Serial)); err != nil {
		return err
	}

	return nil
}
