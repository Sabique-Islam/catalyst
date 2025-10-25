/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	compile "github.com/Sabique-Islam/catalyst/internal/compile"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Build and execute the C/C++ program",
	Long: `Build and execute the C/C++ program. 

If source files are provided, it will build them first and then run the resulting binary.
If no source files are provided, it will try to run the existing binary at bin/project.

Examples:
  catalyst run src/main.c              # Build and run
  catalyst run src/main.c src/utils.c  # Build multiple files and run
  catalyst run                         # Run existing binary`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return compile.RunProject(args)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
