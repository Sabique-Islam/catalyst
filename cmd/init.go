package cmd

import (
	"github.com/Sabique-Islam/catalyst/internal/project"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Catalyst project",
	Long: `Initialize a new Catalyst project with interactive setup.

This command will guide you through setting up a new project configuration
including project name, author, license, and dependencies.

Example:
  catalyst init`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return project.InitializeProject()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
