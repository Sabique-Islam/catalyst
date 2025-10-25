package cmd

import (
	"fmt"

	"github.com/Sabique-Islam/catalyst/internal/compile"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Install dependencies and compile the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1️⃣ Reuse install command logic
		if err := InstallDependencies(); err != nil {
			return err
		}

		// 2️⃣ Compile project
		if err := build.Compile(); err != nil {
			return fmt.Errorf("compilation failed: %w", err)
		}

		fmt.Println("✅ Build complete")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
