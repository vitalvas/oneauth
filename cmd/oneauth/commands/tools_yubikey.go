package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/yubikey"
)

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
