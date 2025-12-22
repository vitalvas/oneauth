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
	"github.com/vitalvas/oneauth/cmd/oneauth/service"
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
		version := buildinfo.Version

		commit := buildinfo.Commit
		if len(commit) > 8 {
			version += "-" + commit[:8]
		}

		log.Printf("OneAuth version: %s", version)

		return nil
	},
	Action: func(c *cli.Context) error {
		group, ctx := errgroup.WithContext(c.Context)

		config, err := config.Load(c.Path("config"))
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		log := logger.New(config.AgentLogPath)

		if config.Keyring.Yubikey.Serial == 0 {
			return fmt.Errorf("yubikey serial is required")
		}

		var agent *sshagent.SSHAgent

		// Track all socket paths for permission fixing
		socketPaths := []string{config.ControlSocketPath}

		// Track all soft agents for shutdown
		softAgents := make(map[string]*sshagent.SoftAgent)

		switch config.Socket.Type {
		case "unix":
			if _, err := os.Stat(config.Socket.Path); err == nil {
				os.Remove(config.Socket.Path)
			}

			if err := tools.MkDir(filepath.Dir(config.Socket.Path), 0700); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			log.WithField("yubikey", config.Keyring.Yubikey.Serial).Println("opening yubikey:", config.Keyring.Yubikey.Serial)

			agent, err = sshagent.New(config.Keyring.Yubikey.Serial, log, config)
			if err != nil {
				return fmt.Errorf("failed to create agent: %w", err)
			}

			if err := service.SetSSHAuthSock(config.Socket.Path); err != nil {
				log.Printf("failed to set SSH_AUTH_SOCK: %v", err)
			}

			socketPaths = append(socketPaths, config.Socket.Path)

			group.Go(func() error {
				return agent.ListenAndServe(ctx, config.Socket.Path)
			})

		case "dummy":
			log.Println("skipping socket creation")

		default:
			return fmt.Errorf("socket type %s is not supported", config.Socket.Type)
		}

		// Start additional soft-key agents
		for name, agentConfig := range config.Agents {
			if _, err := os.Stat(agentConfig.SocketPath); err == nil {
				os.Remove(agentConfig.SocketPath)
			}

			if err := tools.MkDir(filepath.Dir(agentConfig.SocketPath), 0700); err != nil {
				return fmt.Errorf("failed to create directory for agent %s: %w", name, err)
			}

			softAgent := sshagent.NewSoftAgent(name, agentConfig.KeepKeySeconds, log)
			softAgents[name] = softAgent
			socketPaths = append(socketPaths, agentConfig.SocketPath)

			group.Go(func() error {
				return softAgent.ListenAndServe(ctx, agentConfig.SocketPath)
			})
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

			// Shutdown all soft agents
			for _, softAgent := range softAgents {
				softAgent.Shutdown()
			}

			if rpcServer != nil {
				rpcServer.Shutdown()
			}

			return err
		})

		group.Go(func() error {
			return xcmd.PeriodicRun(ctx, func(_ context.Context) error {
				for _, path := range socketPaths {
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
