package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	core "github.com/Sabique-Islam/catalyst/internal/config"
	"github.com/Sabique-Islam/catalyst/internal/fetch"
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
			if strings.HasPrefix(name, ".") || name == "build" || name == "dist" || name == "node_modules" {
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

	return sources, err
}

// InitializeProject runs the interactive project initialization wizard
func InitializeProject() error {
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

		// Detect OS and package manager
		osName := platform.DetectOS()
		pkgManager, err := platform.DetectPackageManager(osName)
		if err != nil {
			return fmt.Errorf("could not detect package manager: %w", err)
		}

		fmt.Printf("Detected OS: %s, Package Manager: %s\n", osName, pkgManager)
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

		for _, abstractName := range abstractDeps {
			realPkgName, found := pkgdb.Translate(abstractName, pkgManager)

			// Add to includes list - ALL headers (both standard and external)
			// Check if it already ends with .h to avoid double extension
			if strings.HasSuffix(abstractName, ".h") {
				includes = append(includes, abstractName)
			} else {
				includes = append(includes, abstractName+".h")
			}

			if !found {
				// Not in package database - likely a project-local header
				fmt.Printf("%s is a local/project header\n", abstractName)
				continue
			}

			// Skip empty package names (standard library headers)
			if realPkgName == "" {
				fmt.Printf("%s is a standard library header (no package needed)\n", abstractName)
				continue
			}

			// Get package names for all major OSes
			darwinPkg, _ := pkgdb.Translate(abstractName, "brew")
			linuxPkg, _ := pkgdb.Translate(abstractName, "apt")
			windowsPkg, _ := pkgdb.Translate(abstractName, "vcpkg")

			if darwinPkg != "" {
				allOsDeps["darwin"] = append(allOsDeps["darwin"], darwinPkg)
			}
			if linuxPkg != "" {
				allOsDeps["linux"] = append(allOsDeps["linux"], linuxPkg)
			}
			if windowsPkg != "" {
				allOsDeps["windows"] = append(allOsDeps["windows"], windowsPkg)
			}

			// Check if already installed on current system
			if platform.IsPackageInstalled(realPkgName, pkgManager) {
				fmt.Printf("%s is already installed\n", realPkgName)
			} else {
				fmt.Printf("%s needs to be installed\n", realPkgName)
			}
		}

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
