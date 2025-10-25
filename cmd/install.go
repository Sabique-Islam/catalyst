package cmd

import (
	"fmt"
	"runtime"

	config "github.com/Sabique-Islam/catalyst/internal/config"
	install "github.com/Sabique-Islam/catalyst/internal/install"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		return InstallDependencies()
	},
}

// InstallDependencies loads the config, gets OS-specific dependencies, and installs them
func InstallDependencies() error {
	// Load catalyst.yml
	cfg, err := config.LoadConfig("catalyst.yml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get dependencies for current OS only
	deps := cfg.GetDependencies() // returns []string
	if len(deps) == 0 {
		fmt.Println("No dependencies to install for this OS.")
		return nil
	}

	fmt.Printf("Installing dependencies for %s: %v\n", runtime.GOOS, deps)
	if err := install.Install(deps); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	fmt.Println("✅ Dependencies installed")
	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)
}
