package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sabique-Islam/catalyst/internal/analyzer"
	core "github.com/Sabique-Islam/catalyst/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	autoMode      bool
	multiTarget   bool
	analyzeReport bool
	dryRun        bool
	interactive   bool
)

// smartInitCmd represents the smart-init command
var smartInitCmd = &cobra.Command{
	Use:   "smart-init",
	Short: "Intelligently initialize Catalyst project with auto-detection",
	Long: `Smart initialization that automatically detects project structure,
dependencies, build targets, and generates appropriate catalyst.yml files.

This command analyzes C project and automatically:
  • Detects build targets (executables with main() functions)
  • Identifies external library dependencies
  • Finds vendored libraries (like cJSON, http-parser)
  • Determines include paths and compiler flags
  • Generates optimized catalyst.yml configurations

Modes:
  --auto          Fully automatic with best guesses (no prompts)
  --interactive   Interactive mode with suggestions (default)
  --dry-run       Show what would be generated without creating files
  --analyze       Show analysis report only

Examples:
  catalyst smart-init                    # Interactive mode
  catalyst smart-init --auto             # Fully automatic
  catalyst smart-init --dry-run          # Preview changes
  catalyst smart-init --analyze          # Analysis report only`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSmartInit()
	},
}

func init() {
	smartInitCmd.Flags().BoolVar(&autoMode, "auto", false, "Fully automatic mode with best guesses")
	smartInitCmd.Flags().BoolVar(&multiTarget, "multi-target", false, "Enable multi-target detection")
	smartInitCmd.Flags().BoolVar(&analyzeReport, "analyze", false, "Show analysis report only")
	smartInitCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be generated without creating files")
	smartInitCmd.Flags().BoolVar(&interactive, "interactive", true, "Interactive mode with suggestions")
	rootCmd.AddCommand(smartInitCmd)
}

func runSmartInit() error {
	fmt.Println("🔍 Analyzing project structure...")
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

	// Show summary
	fmt.Println(scanner.GetSummary())

	// If analyze-only mode, stop here
	if analyzeReport {
		return nil
	}

	// Check if any targets found
	if len(scanner.BuildTargets) == 0 {
		fmt.Println("⚠️  No build targets detected (no main() functions found)")
		fmt.Println("   Run 'catalyst init' for manual setup instead.")
		return nil
	}

	// Generate configurations
	generator := analyzer.NewConfigGenerator(scanner, cwd)
	configs, err := generator.GenerateConfigs()
	if err != nil {
		return fmt.Errorf("failed to generate configs: %w", err)
	}

	// Show generation strategy
	fmt.Println("📝 Configuration Strategy:")
	if len(configs) == 1 {
		fmt.Println("   → Single catalyst.yml (one build target)")
	} else {
		fmt.Println(fmt.Sprintf("   → Separate configs (%d build targets)", len(configs)))
	}
	fmt.Println()

	// Display or create configs
	for configPath, config := range configs {
		fullPath := filepath.Join(cwd, configPath)

		if dryRun {
			// Dry run mode - show what would be created
			fmt.Printf("Would create: %s\n", configPath)
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			yamlData, _ := yaml.Marshal(config)
			fmt.Println(string(yamlData))
			fmt.Println()
		} else {
			// Check if file already exists
			if _, err := os.Stat(fullPath); err == nil {
				if !autoMode {
					fmt.Printf("%s already exists. Overwrite? (y/N): ", configPath)
					var response string
					fmt.Scanln(&response)
					if response != "y" && response != "Y" {
						fmt.Printf("   Skipping %s\n", configPath)
						continue
					}
				} else {
					fmt.Printf("%s already exists, skipping...\n", configPath)
					continue
				}
			}

			// Create the config file
			if err := writeConfig(fullPath, config); err != nil {
				fmt.Printf("Failed to create %s: %v\n", configPath, err)
				continue
			}

			fmt.Printf("Created: %s\n", configPath)
		}
	}

	if !dryRun {
		fmt.Println()
		fmt.Println("✨ Smart initialization complete!")
		fmt.Println()
		fmt.Println("Next steps:")
		if len(configs) == 1 {
			fmt.Println("  catalyst build    # Build the project")
			fmt.Println("  catalyst run      # Build and run")
		} else {
			fmt.Println("  cd <target-dir> && catalyst build")
			for configPath := range configs {
				dir := filepath.Dir(configPath)
				if dir != "." {
					fmt.Printf("  cd %s && catalyst build\n", dir)
				}
			}
		}
	}

	return nil
}

func writeConfig(path string, config *core.Config) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
