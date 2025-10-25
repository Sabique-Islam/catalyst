package cmd

import (
	"fmt"
	"strings"

	"github.com/Sabique-Islam/catalyst/internal/fetch"
	"github.com/Sabique-Islam/catalyst/internal/install"
	"github.com/Sabique-Islam/catalyst/internal/pkgdb"
	"github.com/Sabique-Islam/catalyst/internal/platform"
	"github.com/spf13/cobra"
)

var (
	doctorInstall bool
	doctorDryRun  bool
	doctorVerbose bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose project issues and missing dependencies",
	Long: `Analyze your C project for missing symbols, undefined references, and suggest solutions.
Optionally install suggested dependencies automatically.`,
	RunE: runDoctor,
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorInstall, "install", false, "Automatically install suggested dependencies")
	doctorCmd.Flags().BoolVar(&doctorDryRun, "dry-run", false, "Show what would be installed without actually installing")
	doctorCmd.Flags().BoolVarP(&doctorVerbose, "verbose", "v", false, "Verbose output")
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	fmt.Println("Catalyst Doctor - Project Analysis")
	fmt.Println("=================================")

	// Detect platform information
	osName := platform.DetectOS()
	pkgManager, err := platform.DetectPackageManager(osName)
	if err != nil {
		fmt.Printf("Warning: Could not detect package manager: %v\n", err)
		fmt.Printf("Setup advice:\n%s\n", platform.GetPackageManagerSetupAdvice())
	} else {
		fmt.Printf("Platform: %s (%s)\n", osName, pkgManager)
	}

	// Scan for header dependencies
	fmt.Println("\nHeader Dependency Analysis:")
	fmt.Println("---------------------------")

	headerDeps, err := fetch.ScanDependencies(projectPath)
	if err != nil {
		return fmt.Errorf("failed to scan dependencies: %w", err)
	}

	if len(headerDeps) == 0 {
		fmt.Println("No header dependencies found.")
	} else {
		fmt.Printf("Found %d unique dependencies: %v\n", len(headerDeps), headerDeps)

		// Resolve header dependencies
		var packageSuggestions []string
		for _, dep := range headerDeps {
			if pkg, found := pkgdb.Translate(dep, pkgManager); found && pkg != "" {
				packageSuggestions = append(packageSuggestions, pkg)
			}
		}

		if len(packageSuggestions) > 0 {
			fmt.Printf("Suggested packages: %v\n", packageSuggestions)
		}
	}

	// Scan for missing symbols
	fmt.Println("\nSymbol Linkage Analysis:")
	fmt.Println("------------------------")

	missing, err := fetch.ScanMissingSymbols(projectPath)
	if err != nil {
		fmt.Printf("Could not analyze symbols: %v\n", err)
	} else if len(missing) == 0 {
		fmt.Println("No missing symbols detected!")
	} else {
		fmt.Printf("Found %d groups of missing symbols:\n\n", len(missing))

		var allSuggestedPackages []string

		for i, group := range missing {
			fmt.Printf("%d. Missing symbols (%s):\n", i+1, group.Category)
			symbolNames := fetch.ExtractSymbolNames(group.Symbols)
			for _, symbol := range symbolNames {
				fmt.Printf("   - %s\n", symbol)
			}

			if len(group.SuggestedFiles) > 0 {
				fmt.Printf("   Create these files: %v\n", group.SuggestedFiles)
			}

			if len(group.SuggestedLibs) > 0 {
				fmt.Printf("   Install these libraries: %v\n", group.SuggestedLibs)

				// Resolve library suggestions to actual packages
				for _, lib := range group.SuggestedLibs {
					if pkg, found := pkgdb.Translate(lib, pkgManager); found && pkg != "" {
						allSuggestedPackages = append(allSuggestedPackages, pkg)
					} else if pkg, found := pkgdb.TranslateWithSearch(lib, pkgManager); found {
						allSuggestedPackages = append(allSuggestedPackages, pkg)
					}
				}
			}

			if len(group.PossibleCauses) > 0 {
				fmt.Printf("   Possible solutions:\n")
				for _, cause := range group.PossibleCauses {
					fmt.Printf("      - %s\n", cause)
				}
			}
			fmt.Println()
		}

		// Install dependencies if requested
		if (doctorInstall || doctorDryRun) && len(allSuggestedPackages) > 0 {
			fmt.Println("Dependency Installation:")
			fmt.Println("-----------------------")

			// Remove duplicates
			uniquePackages := removeDuplicates(allSuggestedPackages)

			if doctorDryRun {
				fmt.Printf("Would install %d packages: %v\n", len(uniquePackages), uniquePackages)
			}

			// Install dependencies
			installer, err := install.NewDependencyInstaller(doctorDryRun, doctorVerbose)
			if err != nil {
				fmt.Printf("Error creating installer: %v\n", err)
			} else {
				results, err := installer.InstallBatch(uniquePackages, 3)
				if err != nil {
					fmt.Printf("Error during installation: %v\n", err)
				} else {
					install.PrintResults(results, doctorVerbose)
				}
			}
		}
	}

	// Summary and recommendations
	fmt.Println("\nRecommendations:")
	fmt.Println("----------------")

	if len(missing) > 0 {
		fmt.Println("1. Create missing implementation files listed above")
		fmt.Println("2. Install suggested libraries with 'catalyst doctor --install'")
		fmt.Println("3. Check your build system (Makefile, CMakeLists.txt) for linking flags")
		fmt.Println("4. Ensure all source files are included in compilation")
	} else {
		fmt.Println("1. Run 'catalyst build' to compile your project")
		fmt.Println("2. Use 'catalyst install' to install any remaining dependencies")
	}

	return nil
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(input []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range input {
		item = strings.TrimSpace(item)
		if item != "" && !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
