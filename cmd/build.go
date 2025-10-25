package cmd

import (
	"fmt"

	"github.com/Sabique-Islam/catalyst/internal/build"
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
		// 1️⃣ Install dependencies first (optional)
		if err := InstallDependencies(); err != nil {
			return err
		}

		// 2️⃣ Separate source files from compiler flags
		sourceFiles := []string{}
		flags := []string{}
		for _, arg := range args {
			if len(arg) > 0 && arg[0] == '-' {
				flags = append(flags, arg)
			} else {
				sourceFiles = append(sourceFiles, arg)
			}
		}

		// 3️⃣ Determine output binary
		output := "bin/project"
		if runtime.GOOS == "windows" {
			output += ".exe"
		}

		// 4️⃣ Compile the C/C++ sources
		if err := build.CompileC(sourceFiles, output, flags); err != nil {
			return err
		}

		fmt.Println("✅ Build complete")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
