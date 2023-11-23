package commands

import (
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/go-piv/piv-go/piv"
	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/yubikey"
)

var dataSignCmd = &cli.Command{
	Name:  "sign",
	Usage: "Sign data from stdin",
	Flags: []cli.Flag{
		&cli.Uint64Flag{
			Name:  "serial",
			Usage: "YubiKey serial number",
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

		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read content from stdin: %w", err)
		}

		cert, err := key.GetCertPublicKey(yubikey.SlotKeyRSA.PIVSlot)
		if err != nil {
			return fmt.Errorf("failed to get certificate: %w", err)
		}

		// TODO: add input pin code support
		auth := piv.KeyAuth{}
		privKey, err := key.PrivateKey(yubikey.SlotKeyRSA.PIVSlot, cert, auth)
		if err != nil {
			return fmt.Errorf("failed to get private key: %w", err)
		}

		hasher := sha256.New()

		if _, err := hasher.Write(content); err != nil {
			return fmt.Errorf("failed to write content to hasher: %w", err)
		}

		signer := privKey.(crypto.Signer)

		signature, err := signer.Sign(rand.Reader, hasher.Sum(nil), crypto.SHA256)
		if err != nil {
			return fmt.Errorf("failed to sign content: %w", err)
		}

		signatureStr := base64.StdEncoding.EncodeToString(signature)

		fmt.Println(signatureStr)

		return nil
	},
}
