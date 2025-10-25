/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/Sabique-Islam/catalyst/internal/config"
	"github.com/Sabique-Islam/catalyst/internal/installer"
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build dependancies",
	Long: `installs all dependancies, based on the catalyst.yml file, while auto-detecting OS`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig("catalyst.yml")
		if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
		}

		if err := installer.Install(cfg.Dependencies); err != nil {
				return fmt.Errorf("installation failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
