package cmd

import (
	compile "github.com/Sabique-Islam/catalyst/internal/compile"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Install dependencies and compile C/C++ sources",
	Long: `Reads catalyst.yml and compiles the C/C++ project.

If no catalyst.yml exists, you can pass source files manually.

Examples:
  catalyst build                        # Build from catalyst.yml
  catalyst build src/main.c src/utils.c # Build specific files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return compile.BuildProject(args)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
