package commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/vitalvas/gokit/xcmd"
	"github.com/vitalvas/oneauth/cmd/oneauth/config"
	"github.com/vitalvas/oneauth/cmd/oneauth/rpcserver"
	"github.com/vitalvas/oneauth/cmd/oneauth/sshagent"
	"github.com/vitalvas/oneauth/internal/buildinfo"
	"github.com/vitalvas/oneauth/internal/logger"
	"github.com/vitalvas/oneauth/internal/tools"
	"golang.org/x/sync/errgroup"
)

var agentCmd = &cli.Command{
	Name:        "agent",
	Usage:       "SSH Agent",
	Description: "All configuration options can be set in the config file",
	Before: func(_ *cli.Context) error {
		log.Printf("OneAuth version: %s", buildinfo.FormattedVersion())

		return nil
	},
	Action: func(c *cli.Context) error {
		group, ctx := errgroup.WithContext(c.Context)

		config, err := config.Load(c.Path("config"))
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		log := logger.New(config.AgentLogPath)

		var agent *sshagent.SSHAgent

		switch config.Socket.Type {
		case "unix":
			// Determine YubiKey serial to use
			var yubikeySerial uint32
			if config.Keyring.DisableYubikey {
				log.Println("YubiKey keyring disabled - running without hardware authentication")
				yubikeySerial = 0 // Disabled mode

			} else {
				// YubiKey serial is required for unix socket type when not disabled
				if config.Keyring.Yubikey.Serial == 0 {
					return fmt.Errorf("yubikey serial is required for unix socket type (or set disable_yubikey: true)")
				}
				yubikeySerial = config.Keyring.Yubikey.Serial
				log.WithField("yubikey", yubikeySerial).Println("opening yubikey:", yubikeySerial)
			}

			// Set up socket
			if _, err := os.Stat(config.Socket.Path); err == nil {
				os.Remove(config.Socket.Path)
			}

			if err := tools.MkDir(filepath.Dir(config.Socket.Path), 0700); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			// Create agent
			agent, err = sshagent.New(yubikeySerial, log, config)
			if err != nil {
				return fmt.Errorf("failed to create agent: %w", err)
			}

			group.Go(func() error {
				return agent.ListenAndServe(ctx, config.Socket.Path)
			})

		case "dummy":
			log.Println("skipping socket creation - running in dummy mode")
			// For dummy mode, we don't need YubiKey, so agent remains nil

		default:
			return fmt.Errorf("socket type %s is not supported", config.Socket.Type)
		}

		rpcServer := rpcserver.New(agent, log)

		group.Go(func() error {
			return rpcServer.ListenAndServe(ctx, config.ControlSocketPath)
		})

		group.Go(func() error {
			err := xcmd.WaitInterrupted(ctx)
			log.Println("shutting down agent")

			go func() {
				time.Sleep(10 * time.Second)
				log.Println("shutdown timeout")
				os.Exit(0)
			}()

			if agent != nil {
				agent.Shutdown()
			}

			if rpcServer != nil {
				rpcServer.Shutdown()
			}

			return err
		})

		group.Go(func() error {
			return xcmd.PeriodicRun(ctx, func(_ context.Context) error {
				for _, path := range []string{
					config.Socket.Path,
					config.ControlSocketPath,
				} {
					if stat, err := os.Stat(path); err == nil {
						if perm := stat.Mode().Perm(); perm != 0600 {
							log.Printf("fixing permissions on %s from %d", path, perm)

							if err := os.Chmod(path, 0600); err != nil {
								return err
							}
						}
					}
				}

				return nil
			}, time.Hour)
		})

		return group.Wait()
	},
}
