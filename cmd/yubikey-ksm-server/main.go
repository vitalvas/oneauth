package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vitalvas/oneauth/internal/ksm/server"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "oneauth-yubikey-ksm-server",
		Short: "YubiKey Key Storage Module (KSM) Server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			configPath, _ := cmd.Flags().GetString("config")
			return server.Run(configPath)
		},
	}

	rootCmd.Flags().StringP("config", "c", "", "path to configuration file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
