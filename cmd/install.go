package cmd

import (
	"errors"

	install "github.com/Sabique-Islam/catalyst/internal/install"
	"github.com/spf13/cobra"
)

var (
	resourcesOnly bool
	depsOnly      bool
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dependencies and external resources",
	Long: `Install system dependencies and download external resources defined in catalyst.yml.

Examples:
  catalyst install                     # Install both dependencies and resources
  catalyst install --deps-only         # Install only system dependencies
  catalyst install --resources-only    # Download only external resources`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if resourcesOnly && depsOnly {
			return errors.New("cannot use both --resources-only and --deps-only flags together")
		}

		if resourcesOnly {
			return install.InstallExternalResourcesOnly()
		}

		if depsOnly {
			// Create a version that only installs system dependencies
			return install.InstallSystemDependenciesOnly()
		}

		// Default: install both
		return install.InstallDependencies()
	},
}

func init() {
	installCmd.Flags().BoolVar(&resourcesOnly, "resources-only", false, "Download only external resources (skip system dependencies)")
	installCmd.Flags().BoolVar(&depsOnly, "deps-only", false, "Install only system dependencies (skip external resources)")
	rootCmd.AddCommand(installCmd)
}
