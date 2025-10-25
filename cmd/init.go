package cmd

import (
	"github.com/Sabique-Islam/catalyst/internal/project"
	"github.com/spf13/cobra"
)

var (
	withAnalysis bool
	installDeps  bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Catalyst project",
	Long: `Initialize a new Catalyst project with interactive setup.

This command will guide you through setting up a new project configuration
including project name, author, license, and dependencies.

Options:
  --with-analysis  Include missing symbol analysis
  --install        Automatically install detected dependencies

Example:
  catalyst init
  catalyst init --with-analysis --install`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return project.InitializeProjectWithOptions(withAnalysis, installDeps)
	},
}

func init() {
	initCmd.Flags().BoolVar(&withAnalysis, "with-analysis", false, "Include missing symbol analysis")
	initCmd.Flags().BoolVar(&installDeps, "install", false, "Automatically install detected dependencies")
	rootCmd.AddCommand(initCmd)
}
