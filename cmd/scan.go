package cmd

import (
	"fmt"

	"github.com/Sabique-Islam/catalyst/internal/fetch"
	"github.com/spf13/cobra"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan project for dependencies",
	Long: `Recursively scans all .c and .h files in the current directory
for #include statements to detect external dependencies.

This helps identify which libraries your C project depends on
by analyzing the header files you're including.

Example:
  catalyst scan`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScan()
	},
}

func runScan() error {
	fmt.Println("==============================================")
	fmt.Println("  Catalyst Dependency Scanner                ")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Println("Scanning all .c and .h files in current directory...")
	fmt.Println()

	// Scan the current directory recursively
	deps, err := fetch.ScanDependencies(".")
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if len(deps) == 0 {
		fmt.Println("No external dependencies found (only standard library headers)")
		return nil
	}

	fmt.Printf("Found %d unique dependencies:\n\n", len(deps))
	for i, dep := range deps {
		fmt.Printf("  %d. %s\n", i+1, dep)
	}

	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'catalyst init' to create catalyst.yml")
	fmt.Println("  2. The init wizard will automatically add these")
	fmt.Println("  3. Run 'catalyst install' to install dependencies")
	fmt.Println("  4. Run 'catalyst build' to compile your project")
	fmt.Println()

	return nil
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
