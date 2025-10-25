package cmd

import (
	compile "github.com/Sabique-Islam/catalyst/internal/compile"
	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean build artifacts",
	Long: `Clean build artifacts including compiled binaries and temporary files.

This command removes:
- bin/ directory and all its contents
- Any compiled executables
- Temporary build files

Example:
  catalyst clean`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return compile.CleanProject()
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
