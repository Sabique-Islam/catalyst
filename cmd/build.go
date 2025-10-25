package cmd

import (
	compile "github.com/Sabique-Islam/catalyst/internal/compile"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Install dependencies and compile C/C++ sources",
	Long: `Usage:
  mycli build <source files> [flags]

Example:
  mycli build src/main.c src/utils.c -O2 -Wall`,
	Args: cobra.MinimumNArgs(1), // require at least one source file
	RunE: func(cmd *cobra.Command, args []string) error {
		return compile.BuildProject(args)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
