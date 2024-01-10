package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/gokit/xcmd"
	"github.com/vitalvas/oneauth/cmd/oneauth/config"
	"github.com/vitalvas/oneauth/cmd/oneauth/sshagent"
	"golang.org/x/sync/errgroup"
)

var agentCmd = &cli.Command{
	Name:        "agent",
	Usage:       "SSH Agent",
	Description: "All configuration options can be set in the config file",
	Action: func(c *cli.Context) error {
		group, ctx := errgroup.WithContext(c.Context)

		config, err := config.Load(c.Path("config"))
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if config.Keyring.Yubikey.Serial == 0 {
			return fmt.Errorf("yubikey serial is required")
		}

		var agent *sshagent.SSHAgent

		switch config.Socket.Type {
		case "unix":
			if _, err := os.Stat(config.Socket.Path); err == nil {
				os.Remove(config.Socket.Path)
			}

			if err := os.MkdirAll(filepath.Dir(config.Socket.Path), 0700); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			log.Println("opening yubikey:", config.Keyring.Yubikey.Serial)

			agent, err = sshagent.New(config.Keyring.Yubikey.Serial)
			if err != nil {
				return fmt.Errorf("failed to create agent: %w", err)
			}

			group.Go(func() error {
				return agent.ListenAndServe(ctx, config.Socket.Path)
			})

		case "dummy":
			log.Println("skipping socket creation")

		default:
			return fmt.Errorf("socket type %s is not supported", config.Socket.Type)
		}

		group.Go(func() error {
			err := xcmd.WaitInterrupted(ctx)
			log.Println("shutting down agent")

			if agent != nil {
				agent.Shutdown()
			}

			return err
		})

		return group.Wait()
	},
}
