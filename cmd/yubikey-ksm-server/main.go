package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/vitalvas/oneauth/internal/ksm/server"
	"github.com/vitalvas/oneauth/internal/logger"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "oneauth-yubikey-ksm-server",
		Short: "YubiKey Key Storage Module (KSM) Server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			configPath, _ := cmd.Flags().GetString("config")

			log := logger.New("")

			srv, err := server.New(configPath)
			if err != nil {
				return err
			}
			defer func() {
				if closeErr := srv.Close(); closeErr != nil {
					log.WithError(closeErr).Error("Failed to close server resources")
				}
			}()

			serverErrChan := make(chan error, 1)

			go func() {
				defer close(serverErrChan)
				if err := srv.Start(); err != nil {
					serverErrChan <- fmt.Errorf("server error: %w", err)
				}
			}()

			log.Info("KSM server started")

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			select {
			case err, ok := <-serverErrChan:
				if !ok {
					log.Info("Server shut down normally")
					return nil
				}
				return err
			case <-sigChan:
				log.Info("Shutting down server...")
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()

			if err := srv.Stop(ctx); err != nil {
				return fmt.Errorf("failed to shutdown server: %w", err)
			}

			log.Info("Server shutdown successfully")
			return nil
		},
	}

	rootCmd.Flags().StringP("config", "c", "", "path to configuration file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
