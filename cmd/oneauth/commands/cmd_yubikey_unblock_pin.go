package commands

import (
	"fmt"
	"os"

	"github.com/go-piv/piv-go/v2/piv"
	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/yubikey"
	"golang.org/x/term"
)

var yubikeyUnblockPinCmd = &cli.Command{
	Name:  "unblock-pin",
	Usage: "Unblock the PIN of a YubiKey",
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

		if retries > 0 {
			return fmt.Errorf("PIN is not blocked. Retries left: %d", retries)
		}

		fmt.Print("Enter PUK code: ")
		if err != nil {
			return err
		}

		pukCode, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}

		fmt.Print("\n")

		newPIN, err := readPin()
		if err != nil {
			return err
		}

		if string(pukCode) == newPIN {
			return fmt.Errorf("PIN and PUK codes must be different")
		}

		if newPIN == piv.DefaultPIN {
			return fmt.Errorf("the new PIN can not be the same as the default one")
		}

		if err := key.Unblock(string(pukCode), newPIN); err != nil {
			return fmt.Errorf("failed to unblock PIN: %w", err)
		}

		fmt.Println("PIN successfully unblocked")

		return nil
	},
}
