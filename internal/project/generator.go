package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	core "github.com/Sabique-Islam/catalyst/internal/config"
	"github.com/Sabique-Islam/catalyst/internal/fetch"
	"github.com/Sabique-Islam/catalyst/internal/install"
	"github.com/Sabique-Islam/catalyst/internal/pkgdb"
	"github.com/Sabique-Islam/catalyst/internal/platform"
	"github.com/Sabique-Islam/catalyst/internal/tui"
	"gopkg.in/yaml.v3"
)

type CatalystConfig struct {
	Project struct {
		Name    string `yaml:"name"`
		Created string `yaml:"created"`
	} `yaml:"project"`

	Settings struct {
		Author  string `yaml:"author"`
		License string `yaml:"license"`
	} `yaml:"settings"`
}

func GenerateYAML(projectName string, authorName string, license string) (string, error) {
	cfg := CatalystConfig{}
	cfg.Project.Name = projectName
	cfg.Project.Created = time.Now().Format(time.RFC3339)
	cfg.Settings.Author = authorName
	cfg.Settings.License = license

	out, err := yaml.Marshal(&cfg)
	if err != nil {
		return "", fmt.Errorf("yaml marshal failed: %w", err)
	}
	return string(out), nil
}

// scanSourceFiles finds all .c and .cpp files in the current directory
func scanSourceFiles(dir string) ([]string, error) {
	var sources []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and build directories
		if info.IsDir() {
			name := filepath.Base(path)
			// Don't skip the current directory "." but skip other hidden dirs
			if (strings.HasPrefix(name, ".") && name != ".") || name == "build" || name == "dist" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check for C/C++ source files
		ext := filepath.Ext(path)
		if ext == ".c" || ext == ".cpp" || ext == ".cc" || ext == ".cxx" {
			// Use relative path from current directory
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			sources = append(sources, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// If we found multiple source files, try to filter intelligently
	if len(sources) > 1 {
		sources = filterSourceFiles(sources)
	}

	return sources, nil
}

// filterSourceFiles intelligently filters source files to include the most relevant ones
func filterSourceFiles(allSources []string) []string {
	var mainSources []string
	var srcDirSources []string
	var otherSources []string

	for _, src := range allSources {
		// Categorize source files
		if strings.HasPrefix(src, "src/") {
			srcDirSources = append(srcDirSources, src)
		} else if !strings.Contains(src, "/") || strings.Count(src, "/") == 0 {
			// Files in root directory
			mainSources = append(mainSources, src)
		} else {
			// Files in subdirectories (like arc-server/, examples/, etc.)
			otherSources = append(otherSources, src)
		}
	}

	// Strategy: Include root level C files and src/ directory files
	// This covers the most common project structure
	var result []string

	// Add main sources (root level)
	result = append(result, mainSources...)

	// Add src directory sources
	result = append(result, srcDirSources...)

	// If we have no main sources and src sources, fall back to all sources
	if len(result) == 0 {
		result = allSources
	}

	fmt.Printf("Filtered %d source files from %d total: %v\n", len(result), len(allSources), result)
	return result
}

// getDependencyForOS gets the dependency package name for a specific OS/package manager
// It tries static translation first, then falls back to dynamic search
func getDependencyForOS(abstractName, pkgManager string) string {
	// First try static translation
	if pkg, found := pkgdb.Translate(abstractName, pkgManager); found {
		return pkg
	}

	// If not found, try dynamic search
	if pkg, found := pkgdb.TranslateWithSearch(abstractName, pkgManager); found {
		return pkg
	}

	return ""
}

// resolveDependenciesForOS resolves dependencies for a specific OS with optional interactivity
func resolveDependenciesForOS(dependencies []string, pkgManager string, interactive bool) []string {
	fmt.Printf("\n--- Resolving dependencies for %s ---\n", pkgManager)

	results := pkgdb.BatchSearch(dependencies, pkgManager, interactive)

	var packages []string
	for _, pkg := range results {
		if pkg != "" { // Skip empty packages (standard library)
			packages = append(packages, pkg)
		}
	}

	return packages
}

// resolveDependenciesAutoForOS resolves dependencies automatically without user interaction
func resolveDependenciesAutoForOS(dependencies []string, pkgManager string) []string {
	var packages []string

	for _, dep := range dependencies {
		// Try static first
		if pkg, found := pkgdb.Translate(dep, pkgManager); found {
			if pkg != "" { // Skip empty (standard library) packages
				packages = append(packages, pkg)
			}
			continue
		}

		// Try dynamic search
		if pkg, found := pkgdb.TranslateWithSearch(dep, pkgManager); found {
			packages = append(packages, pkg)
		}
	}

	return packages
}

// InitializeProject runs the interactive project initialization wizard
func InitializeProject() error {
	return InitializeProjectWithOptions(false, false)
}

// InitializeProjectWithOptions runs the project initialization with additional options
func InitializeProjectWithOptions(withAnalysis, installDeps bool) error {
	fmt.Println("==============================================")
	fmt.Println("     Catalyst Project Initialization          ")
	fmt.Println("==============================================")
	fmt.Println()

	// Run the interactive wizard
	config, automate, err := tui.RunInitWizard()
	if err != nil {
		return fmt.Errorf("initialization wizard failed: %w", err)
	}

	// Set metadata
	config.CreatedAt = time.Now().Format(time.RFC3339)

	if automate {
		fmt.Println()
		fmt.Println("Scanning project for dependencies...")

		// If an entry point/source was provided by the wizard, respect it.
		if len(config.Sources) == 0 {
			// Scan for source files
			sources, err := scanSourceFiles(".")
			if err != nil {
				return fmt.Errorf("source file scan failed: %w", err)
			}

			if len(sources) > 0 {
				config.Sources = sources
				fmt.Printf("Found %d source file(s): %v\n", len(sources), sources)
			} else {
				fmt.Println("Warning: No .c or .cpp files found in project")
			}
		} else {
			fmt.Printf("Using specified entry point/source: %v\n", config.Sources)
		}

		// Set default output name if not set
		if config.Output == "" {
			config.Output = config.ProjectName
		}

		// Scan for dependencies
		abstractDeps, err := fetch.ScanDependencies(".")
		if err != nil {
			return fmt.Errorf("dependency scan failed: %w", err)
		}

		if len(abstractDeps) == 0 {
			fmt.Println("No external dependencies found (only standard library headers)")
		} else {
			fmt.Printf("Found %d unique dependencies: %v\n", len(abstractDeps), abstractDeps)
		}

		// Perform symbol analysis if requested
		var symbolDeps []string
		if withAnalysis {
			fmt.Println("Performing missing symbol analysis...")
			missingSymbols, err := fetch.ScanMissingSymbols(".")
			if err != nil {
				fmt.Printf("Warning: Symbol analysis failed: %v\n", err)
			} else if len(missingSymbols) > 0 {
				fmt.Printf("Found %d groups of missing symbols\n", len(missingSymbols))

				for i, group := range missingSymbols {
					fmt.Printf("Group %d (%s): %v\n", i+1, group.Category, fetch.ExtractSymbolNames(group.Symbols))

					if len(group.SuggestedLibs) > 0 {
						fmt.Printf("  Suggested libraries: %v\n", group.SuggestedLibs)
						symbolDeps = append(symbolDeps, group.SuggestedLibs...)
					}

					if len(group.SuggestedFiles) > 0 {
						fmt.Printf("  Missing files: %v\n", group.SuggestedFiles)
					}
				}

				// Add symbol-based dependencies to the main list
				abstractDeps = append(abstractDeps, symbolDeps...)

				// Remove duplicates
				uniqueDeps := make(map[string]bool)
				var finalDeps []string
				for _, dep := range abstractDeps {
					if !uniqueDeps[dep] {
						uniqueDeps[dep] = true
						finalDeps = append(finalDeps, dep)
					}
				}
				abstractDeps = finalDeps

				fmt.Printf("Total dependencies after symbol analysis: %v\n", abstractDeps)
			}
		}

		// Detect OS and package manager
		osName := platform.DetectOS()
		pkgManager, err := platform.DetectPackageManager(osName)
		if err != nil {
			fmt.Printf("Could not detect package manager: %v\n", err)
			fmt.Printf("Setup advice:\n%s\n", platform.GetPackageManagerSetupAdvice())
			return fmt.Errorf("package manager not available")
		}

		fmt.Printf("Detected OS: %s, Package Manager: %s\n", osName, pkgManager)

		// Setup and verify package manager tools
		if err := platform.SetupPackageManager(pkgManager); err != nil {
			fmt.Printf("Warning: %v\n", err)
		}

		fmt.Println()

		// Translate abstract dependencies to real package names
		// Collect dependencies per OS
		// Initialize with all major platforms
		allOsDeps := map[string][]string{
			"darwin":  {},
			"linux":   {},
			"windows": {},
		}
		includes := []string{}

		// Extract resolution preference from config (temporary storage in Author field)
		resolutionMethod := config.Author
		if resolutionMethod == "" {
			resolutionMethod = "auto" // Default to automatic
		}

		fmt.Printf("Using %s dependency resolution...\n\n", resolutionMethod)

		// Resolve dependencies based on preference
		if resolutionMethod == "interactive" {
			// Interactive mode - let user choose for each OS
			allOsDeps["darwin"] = resolveDependenciesForOS(abstractDeps, "brew", true)
			allOsDeps["linux"] = resolveDependenciesForOS(abstractDeps, "apt", true)
			allOsDeps["windows"] = resolveDependenciesForOS(abstractDeps, "vcpkg", true)
		} else if resolutionMethod == "database" {
			// Database only mode
			allOsDeps["darwin"] = resolveDependenciesForOS(abstractDeps, "brew", false)
			allOsDeps["linux"] = resolveDependenciesForOS(abstractDeps, "apt", false)
			allOsDeps["windows"] = resolveDependenciesForOS(abstractDeps, "vcpkg", false)
		} else {
			// Auto mode - use enhanced resolution without interaction
			fmt.Println("Automatically resolving dependencies for all platforms...")
			allOsDeps["darwin"] = resolveDependenciesAutoForOS(abstractDeps, "brew")
			allOsDeps["linux"] = resolveDependenciesAutoForOS(abstractDeps, "apt")
			allOsDeps["windows"] = resolveDependenciesAutoForOS(abstractDeps, "vcpkg")
		}

		// Build includes list
		for _, abstractName := range abstractDeps {
			// Add to includes list - ALL headers (both standard and external)
			// Check if it already ends with .h to avoid double extension
			if strings.HasSuffix(abstractName, ".h") {
				includes = append(includes, abstractName)
			} else {
				includes = append(includes, abstractName+".h")
			}
		}

		// Check what's already installed on current system (for info only)
		fmt.Println("\nChecking current system installation status...")
		for _, abstractName := range abstractDeps {
			if realPkgName, found := pkgdb.Translate(abstractName, pkgManager); found && realPkgName != "" {
				if platform.IsPackageInstalled(realPkgName, pkgManager) {
					fmt.Printf("✓ %s (%s) is already installed\n", abstractName, realPkgName)
				} else {
					fmt.Printf("✗ %s (%s) needs to be installed\n", abstractName, realPkgName)
				}
			}
		}

		// Clear the temporary resolution preference from Author field
		config.Author = ""

		// Remove duplicates from each OS dependency list
		for os, deps := range allOsDeps {
			uniqueDeps := make(map[string]bool)
			uniqueList := []string{}
			for _, dep := range deps {
				if !uniqueDeps[dep] {
					uniqueDeps[dep] = true
					uniqueList = append(uniqueList, dep)
				}
			}
			allOsDeps[os] = uniqueList
		}

		// Populate config with dependencies for all OSes
		// allOsDeps is always initialized with all platforms
		config.Dependencies = allOsDeps

		// Add includes to config
		if len(includes) > 0 {
			config.Includes = includes
		}

		// Install dependencies if requested
		if installDeps {
			fmt.Println("\nInstalling dependencies...")
			currentOsDeps := allOsDeps[osName]
			if len(currentOsDeps) > 0 {
				installer, err := install.NewDependencyInstaller(false, true)
				if err != nil {
					fmt.Printf("Warning: Could not create installer: %v\n", err)
				} else {
					results, err := installer.InstallBatch(currentOsDeps, 3)
					if err != nil {
						fmt.Printf("Error during installation: %v\n", err)
					} else {
						install.PrintResults(results, true)
					}
				}
			} else {
				fmt.Println("No dependencies to install for current platform.")
			}
		}

		// Save config using standard method (now includes has its own field)
		if err := saveConfig(config, "catalyst.yml"); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

	} else {
		// Manual mode - just save basic config
		fmt.Println()
		fmt.Println("Creating basic catalyst.yml template...")
		fmt.Println("   You'll need to manually add dependencies and includes.")

		// Initialize empty dependency structure for all major platforms
		// This ensures the dependencies section appears in the YAML file
		config.Dependencies = map[string][]string{
			"darwin":  {},
			"linux":   {},
			"windows": {},
		}

		if err := saveConfig(config, "catalyst.yml"); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}
	}

	fmt.Println()
	fmt.Println("Project initialized successfully!")
	fmt.Printf("Configuration saved to: catalyst.yml\n")

	if automate {
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  1. Review catalyst.yml")
		fmt.Println("  2. Run 'catalyst install' to install dependencies")
		fmt.Println("  3. Run 'catalyst build' to compile your project")
	}

	return nil
}

// saveConfig writes the config to a YAML file
func saveConfig(cfg *core.Config, filename string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
