/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	install "github.com/Sabique-Islam/catalyst/internal/install"
	config "github.com/Sabique-Islam/catalyst/internal/config"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dependencies required by the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.LoadConfig("catalyst.yml")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Install dependencies using internal/install (auto-detects OS)
		if err := install.Install(cfg.Dependencies); err != nil {
			return fmt.Errorf("installation failed: %w", err)
		}

		fmt.Println("install complete")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
