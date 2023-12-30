package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-piv/piv-go/piv"
	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/yubikey"
	"golang.org/x/term"
)

var yubikeyChangePinCmd = &cli.Command{
	Name:  "change-pin",
	Usage: "Change the PIN of a YubiKey",
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

		retries, err := key.Retries()
		if err != nil {
			return fmt.Errorf("failed to get PIN retries: %w", err)
		}

		if retries == 0 {
			return errors.New("PIN is blocked. Unblock it with PUK code")
		}

		fmt.Print("Enter current PIN: ")
		currentPIN, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}

		fmt.Print("\n")

		if !yubikey.ValidatePin(string(currentPIN)) {
			return fmt.Errorf("invalid PIN")
		}

		if err := key.VerifyPIN(string(currentPIN)); err != nil {
			return fmt.Errorf("failed to verify current PIN: %w", err)
		}

		newPIN, err := readPin()
		if err != nil {
			return err
		}

		if string(currentPIN) == newPIN {
			return fmt.Errorf("the new PIN can not be the same as the current one")
		}

		if err := key.SetPIN(string(currentPIN), newPIN); err != nil {
			return fmt.Errorf("failed to change PIN: %w", err)
		}

		fmt.Println("PIN changed successfully")

		return nil
	},
}

func readPin() (string, error) {
	fmt.Println("The PIN code can consist of 6-8 digits (0-9)")

	fmt.Print("Enter new PIN: ")
	newPIN, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	if !yubikey.ValidatePin(string(newPIN)) {
		return "", fmt.Errorf("invalid PIN")
	}

	if string(newPIN) == piv.DefaultPIN {
		return "", fmt.Errorf("the new PIN can not be the same as the default one")
	}

	fmt.Print("\nRepeat new PIN: ")

	repeatNewPIN, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	fmt.Print("\n")

	if string(newPIN) != string(repeatNewPIN) {
		return "", fmt.Errorf("the new PINs do not match")
	}

	return string(newPIN), nil
}
