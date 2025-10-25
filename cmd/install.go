package cmd

import (
	"fmt"

	install "github.com/Sabique-Islam/catalyst/internal/install"
	config "github.com/Sabique-Islam/catalyst/internal/config"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		return InstallDependencies()
	},
}

func InstallDependencies() error {
	cfg, err := config.LoadConfig("catalyst.yml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	deps := cfg.GetDependencies()
	if len(deps) == 0 {
		fmt.Println("No dependencies to install for this OS.")
		return nil
	}

	fmt.Printf("Installing dependencies for %s: %v\n", runtime.GOOS, deps)
	if err := build.Install(deps); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	fmt.Println("✅ Dependencies installed")
	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)
}
