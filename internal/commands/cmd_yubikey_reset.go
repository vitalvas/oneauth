package commands

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/yubikey"
)

var yubikeyResetCmd = &cli.Command{
	Name:  "reset",
	Usage: "Reset Yubikey to default settings",
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

		if err := key.ResetToDefault(); err != nil {
			return err
		}

		fmt.Println("Done")

		return nil
	},
}
