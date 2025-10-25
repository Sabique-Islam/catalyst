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
		realDeps := []string{}
		includes := []string{}

		for _, abstractName := range abstractDeps {
			realPkgName, found := pkgdb.Translate(abstractName, pkgManager)
			if !found {
				// Not in package database - likely a project-local header, skip it
				continue
			}

			// Add to includes list (this is a real system header)
			includes = append(includes, abstractName+".h")

			// Skip empty package names (standard library headers)
			if realPkgName == "" {
				fmt.Printf("%s is a standard library header (no package needed)\n", abstractName)
				continue
			}

			// Check if already installed
			if platform.IsPackageInstalled(realPkgName, pkgManager) {
				fmt.Printf("%s is already installed\n", realPkgName)
			} else {
				fmt.Printf("%s needs to be installed\n", realPkgName)
			}

			realDeps = append(realDeps, realPkgName)
		}

		// Populate config with dependencies
		if len(realDeps) > 0 {
			config.Dependencies = map[string][]string{
				osName: realDeps,
			}
		}

		// Add includes section (write manually to YAML)
		if len(includes) > 0 {
			if err := saveConfigWithIncludes(config, "catalyst.yml", includes); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
		} else {
			if err := saveConfig(config, "catalyst.yml"); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
		}

	} else {
		// Manual mode - just save basic config
		fmt.Println()
		fmt.Println("Creating basic catalyst.yml template...")
		fmt.Println("   You'll need to manually add dependencies and includes.")

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

// saveConfigWithIncludes writes the config with a custom includes section
func saveConfigWithIncludes(cfg *core.Config, filename string, includes []string) error {
	// Create a custom structure to include the "includes" field
	type ConfigWithIncludes struct {
		ProjectName  string              `yaml:"project_name"`
		Sources      []string            `yaml:"sources,omitempty"`
		Output       string              `yaml:"output,omitempty"`
		Flags        []string            `yaml:"flags,omitempty"`
		Dependencies map[string][]string `yaml:"dependencies,omitempty"`
		Includes     []string            `yaml:"includes,omitempty"`
		Resources    []string            `yaml:"resources,omitempty"`
	}

	customCfg := ConfigWithIncludes{
		ProjectName:  cfg.ProjectName,
		Sources:      cfg.Sources,
		Output:       cfg.Output,
		Flags:        cfg.Flags,
		Dependencies: cfg.Dependencies,
		Includes:     includes,
	}

	// Convert Resources to string array if needed
	if len(cfg.Resources) > 0 {
		resources := make([]string, len(cfg.Resources))
		for i, r := range cfg.Resources {
			if r.Path != "" {
				resources[i] = r.Path
			} else {
				resources[i] = r.URL
			}
		}
		customCfg.Resources = resources
	}

	data, err := yaml.Marshal(customCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
