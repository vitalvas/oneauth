package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/buildinfo"
	"github.com/vitalvas/oneauth/internal/updates"
)

var updateCmd = &cli.Command{
	Name:  "update",
	Usage: "update oneauth",
	Action: func(c *cli.Context) error {
		manifest, err := updates.Check("oneauth", buildinfo.Version)
		if err != nil {
			if err == updates.ErrNoUpdateAvailable {
				fmt.Println("No update available")
				return nil
			}

			return fmt.Errorf("failed to check for updates: %w", err)
		}

		fmt.Printf("New version available: (current: %s) %s\n", buildinfo.Version, manifest.Version)

		return nil
	},
}
