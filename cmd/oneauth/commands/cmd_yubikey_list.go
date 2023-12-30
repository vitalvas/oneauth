package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/tools"
	"github.com/vitalvas/oneauth/internal/yubikey"
)

var yubikeyListCmd = &cli.Command{
	Name:  "list",
	Usage: "List Yubikeys",
	Action: func(c *cli.Context) error {
		cards, err := yubikey.Cards()
		if err != nil {
			return err
		}

		for idx, card := range cards {
			yk, err := yubikey.Open(card)
			if err != nil {
				return err
			}

			if idx > 0 {
				fmt.Println(strings.Repeat("-", 24))
			}

			fmt.Println(card.Name)
			fmt.Println(" - Serial:", card.Serial)
			fmt.Println(" - Version:", card.Version)
			fmt.Println(" - Keys:")

			keys, err := yk.ListKeys(yubikey.AllSlots...)
			if err != nil {
				return err
			}

			if len(keys) == 0 {
				fmt.Println("   - no keys")
			}

			for _, key := range keys {
				fmt.Printf("   - %s | %s:\n", key.Slot.PIVSlot.String(), key.Subject.CommonName)
				fmt.Printf("     - created: %s expires: %s\n", key.NotBefore.Local().Format(time.RFC3339), key.NotAfter.Local().Format(time.RFC3339))

				if certSSHKey, err := tools.GetSSHPublicKey(key.PublicKey); err == nil {
					certSSHKeyStr := strings.TrimSpace(string(certSSHKey))
					fmt.Printf("     - SSH: %s\n", certSSHKeyStr)
				}
			}

			yk.Close()
		}

		return nil
	},
}
