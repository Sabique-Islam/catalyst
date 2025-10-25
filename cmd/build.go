/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	core "github.com/Sabique-Islam/catalyst/internal/config"
	"github.com/Sabique-Islam/catalyst/internal/build"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Install dependencies and build the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.LoadConfig("catalyst.yml")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get OS-specific dependencies
		deps := cfg.GetDependencies()

		// Install dependencies using internal/build
		if err := build.Install(deps); err != nil {
			return fmt.Errorf("installation failed: %w", err)
		}

		fmt.Println("✅ Build complete")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
