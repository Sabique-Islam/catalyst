package cmd

import (
	"fmt"
	"os"

	"github.com/Sabique-Islam/catalyst/internal/analyzer"
	"github.com/spf13/cobra"
)

var (
	verboseAnalysis bool
	showDeps        bool
	showTargets     bool
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze project structure and dependencies",
	Long: `Analyze your C/C++ project to understand its structure,
dependencies, and build requirements without modifying anything.

This command provides detailed insights about:
  • Source and header files
  • Build targets (executables)
  • External library dependencies
  • Vendored/bundled libraries
  • Include relationships

Use this before 'smart-init' to understand what will be generated.

Examples:
  catalyst analyze                 # Basic analysis
  catalyst analyze --verbose       # Detailed analysis
  catalyst analyze --show-deps     # Focus on dependencies
  catalyst analyze --show-targets  # Focus on build targets`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAnalyze()
	},
}

func init() {
	analyzeCmd.Flags().BoolVarP(&verboseAnalysis, "verbose", "v", false, "Show detailed analysis")
	analyzeCmd.Flags().BoolVar(&showDeps, "show-deps", false, "Focus on dependencies")
	analyzeCmd.Flags().BoolVar(&showTargets, "show-targets", false, "Focus on build targets")
	rootCmd.AddCommand(analyzeCmd)
}

func runAnalyze() error {
	fmt.Println("🔍 Analyzing project...")
	fmt.Println()

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Create and run scanner
	scanner := analyzer.NewProjectScanner(cwd)
	if err := scanner.ScanProject(); err != nil {
		return fmt.Errorf("failed to scan project: %w", err)
	}

	// Show basic summary (always)
	fmt.Println(scanner.GetSummary())

	// Verbose mode - show more details
	if verboseAnalysis || showTargets {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("  Detailed Build Target Analysis")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		for i, target := range scanner.BuildTargets {
			fmt.Printf("%d. Target: %s\n", i+1, target.Name)
			fmt.Printf("   Type: %s\n", target.Type)
			fmt.Printf("   Entry Point: %s\n", target.EntryPoint)
			fmt.Printf("   Directory: %s\n", target.Directory)
			fmt.Println("   Source Files:")
			for _, src := range target.SourceFiles {
				fmt.Printf("     • %s\n", src)
			}
			fmt.Println()
		}
	}

	if verboseAnalysis || showDeps {
		if len(scanner.ExternalLibs) > 0 {
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("  External Dependencies Detail")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println()

			for _, lib := range scanner.ExternalLibs {
				fmt.Printf("Library: %s\n", lib.Name)
				fmt.Printf("  Header: %s\n", lib.HeaderName)
				fmt.Printf("  Linker Flag: %s\n", lib.LinkerFlag)
				if lib.PkgConfig != "" {
					fmt.Printf("  pkg-config: %s\n", lib.PkgConfig)
				}
				fmt.Println("  Platform Packages:")
				for platform, pkg := range lib.Platforms {
					if pkg.PackageName != "" {
						fmt.Printf("    %s: %s\n", platform, pkg.PackageName)
					}
				}
				fmt.Println()
			}
		}

		if len(scanner.VendoredLibs) > 0 {
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("  Vendored Libraries Detail")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println()

			for _, lib := range scanner.VendoredLibs {
				fmt.Printf("Library: %s\n", lib.Name)
				fmt.Printf("  Location: %s/\n", lib.Path)
				fmt.Println("  Source Files:")
				for _, src := range lib.SourceFiles {
					fmt.Printf("    • %s\n", src)
				}
				fmt.Println("  Header Files:")
				for _, hdr := range lib.HeaderFiles {
					fmt.Printf("    • %s\n", hdr)
				}
				fmt.Println()
			}
		}
	}

	// Show recommendations
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Recommendations")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if len(scanner.BuildTargets) == 0 {
		fmt.Println("  No build targets detected")
		fmt.Println("   → No main() functions found in source files")
		fmt.Println("   → This might be a library project")
		fmt.Println("   → Use 'catalyst init' for manual setup")
	} else if len(scanner.BuildTargets) == 1 {
		fmt.Println(" Single build target detected")
		fmt.Println("   → Use 'catalyst smart-init' to auto-generate config")
	} else {
		fmt.Println(" Multiple build targets detected")
		fmt.Println("   → Use 'catalyst smart-init --multi-target'")
		fmt.Println("   → Will create separate catalyst.yml for each target")
	}

	if len(scanner.ExternalLibs) > 0 {
		fmt.Println()
		fmt.Printf(" %d external dependencies detected\n", len(scanner.ExternalLibs))
		fmt.Println("   → smart-init will auto-configure these")
	}

	if len(scanner.VendoredLibs) > 0 {
		fmt.Println()
		fmt.Printf(" %d vendored libraries detected\n", len(scanner.VendoredLibs))
		fmt.Println("   → smart-init will include these in build")
	}

	return nil
}
