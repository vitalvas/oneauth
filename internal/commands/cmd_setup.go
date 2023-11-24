package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-piv/piv-go/piv"
	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/certgen"
	"github.com/vitalvas/oneauth/internal/tools"
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
		&cli.Uint64Flag{
			Name:  "valid-days",
			Usage: "Number of days the insecure keys will be valid",
			Value: 3650,
		},
		&cli.Uint64Flag{
			Name:  "ecc-bits",
			Usage: "Number of bits for the insecure ECC keys. Supported values are 256 and 384",
			Value: 256,
		},
		&cli.StringFlag{
			Name:  "touch-policy",
			Usage: "Touch policy for the insecure keys. Supported values are cached, always and never",
			Value: "cached",
		},
	},
	Before: selectYubiKey,
	Action: func(c *cli.Context) error {
		var afterLines []string

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

		if validDays := c.Uint64("valid-days"); validDays > 0 {
			username := c.String("username")

			var touchPolicy piv.TouchPolicy
			switch c.String("touch-policy") {
			case "never":
				touchPolicy = piv.TouchPolicyNever

			case "cached":
				touchPolicy = piv.TouchPolicyCached

			case "always":
				touchPolicy = piv.TouchPolicyAlways

			default:
				return fmt.Errorf("unsupported touch policy: %s", c.String("touch-policy"))
			}

			key.GenCertificate(yubikey.MustSlotFromKeyID(yubikey.SlotKeyRSAID), newPIN, yubikey.CertRequest{
				CommonName: certgen.GenCommonName(username, "insecure-rsa"),
				Days:       int(validDays),
				Key: piv.Key{
					Algorithm:   piv.AlgorithmRSA2048,
					PINPolicy:   piv.PINPolicyNever,
					TouchPolicy: touchPolicy,
				},
			})

			var eccAlgo piv.Algorithm

			switch c.Uint64("ecc-bits") {
			case 256:
				eccAlgo = piv.AlgorithmEC256

			case 384:
				eccAlgo = piv.AlgorithmEC384

			default:
				return fmt.Errorf("unsupported ECC bits: %d", c.Uint64("ecc-bits"))
			}

			key.GenCertificate(yubikey.MustSlotFromKeyID(yubikey.SlotKeyECDSAID), newPIN, yubikey.CertRequest{
				CommonName: certgen.GenCommonName(username, "insecure-ecdsa"),
				Days:       int(validDays),
				Key: piv.Key{
					Algorithm:   eccAlgo,
					PINPolicy:   piv.PINPolicyNever,
					TouchPolicy: touchPolicy,
				},
			})

			keys, err := key.ListKeys(yubikey.SlotKeyRSA, yubikey.SlotKeyECDSA)
			if err != nil {
				return fmt.Errorf("failed to list keys: %w", err)
			}

			for _, key := range keys {
				if certSSHKey, err := tools.GetSSHPublicKey(key.PublicKey); err == nil {
					certSSHKeyStr := strings.TrimSpace(string(certSSHKey))
					afterLines = append(afterLines, fmt.Sprintf("Insecure SSH: %s", certSSHKeyStr))
				}
			}

		} else {
			fmt.Println("Skipping insecure keys generation")
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
