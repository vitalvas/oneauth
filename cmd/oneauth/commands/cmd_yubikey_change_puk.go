package commands

import (
	"fmt"
	"os"

	"github.com/go-piv/piv-go/piv"
	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/yubikey"
	"golang.org/x/term"
)

var yubikeyChangePukCmd = &cli.Command{
	Name:  "change-puk",
	Usage: "Change the PUK of a YubiKey",
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

		fmt.Printf("YubiKey with serial %d is connected\n", serial)

		fmt.Print("Enter current PUK: ")
		currentPUK, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}

		fmt.Print("\n")

		if !yubikey.ValidatePuk(string(currentPUK)) {
			return fmt.Errorf("invalid PUK")
		}

		newPUK, err := readPuk()
		if err != nil {
			return err
		}

		if string(currentPUK) == newPUK {
			return fmt.Errorf("the new PUK can not be the same as the current one")
		}

		if err := key.SetPUK(string(currentPUK), newPUK); err != nil {
			return fmt.Errorf("failed to change PUK: %w", err)
		}

		fmt.Println("PUK changed successfully")

		return nil
	},
}

func readPuk() (string, error) {
	fmt.Println("The PUK code can consist of 8 digits (0-9)")

	fmt.Print("Enter new PUK: ")
	newPUK, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	if !yubikey.ValidatePuk(string(newPUK)) {
		return "", fmt.Errorf("invalid PUK")
	}

	if string(newPUK) == piv.DefaultPUK {
		return "", fmt.Errorf("the new PUK can not be the same as the default one")
	}

	fmt.Print("\nRepeat new PUK: ")

	repeatNewPUK, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	fmt.Print("\n")

	if string(newPUK) != string(repeatNewPUK) {
		return "", fmt.Errorf("the new PUKs do not match")
	}

	return string(newPUK), nil
}
