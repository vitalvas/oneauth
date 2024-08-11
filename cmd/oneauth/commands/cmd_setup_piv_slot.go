package commands

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/go-piv/piv-go/piv"
	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/certgen"
	"github.com/vitalvas/oneauth/internal/keyring"
	"github.com/vitalvas/oneauth/internal/yubikey"
)

var setupPivSlotCmd = &cli.Command{
	Name:  "piv-slot",
	Usage: "Setup a new YubiKey PIV slot",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:     "confirm",
			Usage:    "Confirm the setup (piv slot will be wiped)",
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
			Name:     "slot",
			Usage:    "PIV slot to setup (0x82-0x95)",
			Required: true,
			Action: func(_ *cli.Context, data string) error {
				if !slices.Contains(yubikey.PIVSlots, strings.TrimLeft(data, "0x")) {
					return fmt.Errorf("unsupported slot: %s", data)
				}

				return nil
			},
		},
		&cli.StringFlag{
			Name:  "username",
			Value: "oneauth",
		},
		&cli.Uint64Flag{
			Name:  "valid-days",
			Usage: "Number of days the insecure keys will be valid",
			Value: 3650,
			Action: func(_ *cli.Context, data uint64) error {
				if data == 0 {
					return fmt.Errorf("valid-days is required")
				}

				return nil
			},
		},
		&cli.StringFlag{
			Name:  "key-type",
			Usage: "Key type for the insecure keys. Supported values are rsa2048, eccp256 and eccp384",
			Value: "eccp256",
			Action: func(_ *cli.Context, data string) error {
				if !slices.Contains([]string{"rsa2048", "eccp256", "eccp384"}, data) {
					return fmt.Errorf("unsupported key type: %s", data)
				}

				return nil
			},
		},
		&cli.StringFlag{
			Name:  "touch-policy",
			Usage: "Touch policy for the insecure keys. Supported values are cached, always and never",
			Value: "cached",
			Action: func(_ *cli.Context, data string) error {
				if _, ok := yubikey.MapTouchPolicy(data); ok {
					return nil
				}

				return fmt.Errorf("unsupported touch policy: %s", data)
			},
		},
		&cli.StringFlag{
			Name:  "pin-policy",
			Usage: "PIN policy for the insecure keys. Supported values are once, always and never",
			Value: "once",
			Action: func(_ *cli.Context, data string) error {
				if _, ok := yubikey.MapPINPolicy(data); ok {
					return nil
				}

				return fmt.Errorf("unsupported PIN policy: %s", data)
			},
		},
	},
	Before: selectYubiKey,
	Action: func(c *cli.Context) error {
		var afterLines []string

		serial := uint32(c.Uint64("serial"))
		if serial == 0 {
			return fmt.Errorf("serial is required")
		}

		key, err := yubikey.OpenBySerial(serial)
		if err != nil {
			return err
		}

		defer key.Close()

		pivSlot := yubikey.MustSlotFromKeyID(uint32(c.Uint("slot")))

		fmt.Println("Setup a PIV slot on YubiKey:", pivSlot.String())

		fmt.Println("Serial:", serial)

		if wait := c.Uint64("wait"); wait > 0 {
			fmt.Printf("Waiting %d seconds for cancel...\n", wait)
			time.Sleep(time.Duration(wait) * time.Second)
		}

		username := c.String("username")

		validDays := c.Uint64("valid-days")

		var touchPolicy piv.TouchPolicy
		if policy, ok := yubikey.MapTouchPolicy(c.String("touch-policy")); ok {
			touchPolicy = policy
		}

		var pinPolicy piv.PINPolicy
		if policy, ok := yubikey.MapPINPolicy(c.String("pin-policy")); ok {
			pinPolicy = policy
		}

		yubikeyPIN, err := keyring.Get(keyring.GetYubikeyAccount(serial, "pin"))
		if err != nil {
			return fmt.Errorf("failed to get YubiKey PIN: %w", err)
		}

		switch c.String("key-type") {
		case "rsa2048":
			key.GenCertificate(pivSlot, yubikeyPIN, yubikey.CertRequest{
				CommonName: certgen.GenCommonName(username, "insecure-rsa"),
				Days:       int(validDays),
				Key: piv.Key{
					Algorithm:   piv.AlgorithmRSA2048,
					PINPolicy:   pinPolicy,
					TouchPolicy: touchPolicy,
				},
			})

		case "eccp256", "eccp384":
			// default to eccp256
			eccAlgo := piv.AlgorithmEC256

			if c.String("key-type") == "eccp384" {
				eccAlgo = piv.AlgorithmEC384
			}

			key.GenCertificate(pivSlot, yubikeyPIN, yubikey.CertRequest{
				CommonName: certgen.GenCommonName(username, "insecure-ecdsa"),
				Days:       int(validDays),
				Key: piv.Key{
					Algorithm:   eccAlgo,
					PINPolicy:   pinPolicy,
					TouchPolicy: touchPolicy,
				},
			})
		}

		fmt.Println("Done")

		if len(afterLines) > 0 {
			fmt.Println(strings.Repeat("-", 60))
		}

		for _, line := range afterLines {
			fmt.Println("-:", line)
		}

		return nil
	},
}
