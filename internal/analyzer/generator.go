package analyzer

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	core "github.com/Sabique-Islam/catalyst/internal/config"
)

// ConfigGenerator generates catalyst.yml configurations from scan results
type ConfigGenerator struct {
	Scanner    *ProjectScanner
	ProjectDir string
}

// NewConfigGenerator creates a new config generator
func NewConfigGenerator(scanner *ProjectScanner, projectDir string) *ConfigGenerator {
	return &ConfigGenerator{
		Scanner:    scanner,
		ProjectDir: projectDir,
	}
}

// GenerateConfigs generates catalyst.yml configurations for detected targets
func (cg *ConfigGenerator) GenerateConfigs() (map[string]*core.Config, error) {
	configs := make(map[string]*core.Config)

	if len(cg.Scanner.BuildTargets) == 0 {
		return nil, fmt.Errorf("no build targets detected")
	}

	// Decide strategy: separate configs for each target
	for _, target := range cg.Scanner.BuildTargets {
		config := cg.generateConfigForTarget(target)

		// Determine config file path
		var configPath string
		if target.Directory != "." && target.Directory != "" {
			configPath = filepath.Join(target.Directory, "catalyst.yml")
		} else {
			configPath = "catalyst.yml"
		}

		configs[configPath] = config
	}

	return configs, nil
}

// generateConfigForTarget generates a config for a specific build target
func (cg *ConfigGenerator) generateConfigForTarget(target BuildTarget) *core.Config {
	config := &core.Config{
		ProjectName:  target.Name,
		Sources:      target.SourceFiles,
		Output:       target.Name,
		Dependencies: make(map[string][]string),
		Flags:        []string{},
		Includes:     []string{},
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	// Add compiler flags
	config.Flags = append(config.Flags, "-Wall", "-Wextra")

	// Determine if C or C++
	isCPP := false
	for _, src := range target.SourceFiles {
		ext := filepath.Ext(src)
		if ext == ".cpp" || ext == ".cc" || ext == ".cxx" {
			isCPP = true
			break
		}
	}

	if !isCPP {
		config.Flags = append(config.Flags, "-std=c99", "-g")
	}

	// Add include paths
	includePaths := cg.collectIncludePaths(target)
	for _, incPath := range includePaths {
		config.Flags = append(config.Flags, "-I"+incPath)
	}

	// Add vendored library sources to target if they're in the same directory tree
	for _, vlib := range cg.Scanner.VendoredLibs {
		if cg.isLibraryUsedByTarget(target, vlib) {
			// Add vendored sources
			for _, src := range vlib.SourceFiles {
				// Make path relative to target directory if needed
				relSrc := cg.makeRelativeToTarget(src, target.Directory)
				if !contains(config.Sources, relSrc) {
					config.Sources = append(config.Sources, relSrc)
				}
			}

			// Add vendored include path
			vlibIncPath := cg.makeRelativeToTarget(vlib.Path, target.Directory)
			config.Flags = append(config.Flags, "-I"+vlibIncPath)
		}
	}

	// Add external library dependencies
	externalLibs := cg.getExternalLibsForTarget(target)

	// Initialize platform dependencies
	config.Dependencies["darwin"] = []string{}
	config.Dependencies["linux"] = []string{}
	config.Dependencies["windows"] = []string{}

	for _, lib := range externalLibs {
		// Add platform-specific dependencies
		for platform, pkg := range lib.Platforms {
			if pkg.PackageName != "" {
				config.Dependencies[platform] = append(config.Dependencies[platform], pkg.PackageName)
			}
		}

		// Add linker flags
		if lib.LinkerFlag != "" {
			flags := strings.Fields(lib.LinkerFlag)
			config.Flags = append(config.Flags, flags...)
		}

		// Add platform-specific include/lib paths (for macOS)
		if runtime.GOOS == "darwin" {
			if pkg, ok := lib.Platforms["darwin"]; ok {
				if pkg.IncludePath != "" {
					config.Flags = append(config.Flags, "-I"+pkg.IncludePath)
				}
				if pkg.LibPath != "" {
					config.Flags = append(config.Flags, "-L"+pkg.LibPath)
				}
			}
		}
	}

	// Add math library if needed
	if !contains(config.Flags, "-lm") {
		config.Flags = append(config.Flags, "-lm")
	}

	// Collect all includes for documentation
	config.Includes = cg.collectAllIncludes(target)

	return config
}

// collectIncludePaths collects include directory paths for a target
func (cg *ConfigGenerator) collectIncludePaths(target BuildTarget) []string {
	paths := make(map[string]bool)

	// Check for common include directories
	commonPaths := []string{"include", "inc", "headers"}
	for _, p := range commonPaths {
		var checkPath string
		if target.Directory != "." && target.Directory != "" {
			checkPath = filepath.Join(cg.ProjectDir, target.Directory, p)
		} else {
			checkPath = filepath.Join(cg.ProjectDir, p)
		}

		if exists(checkPath) {
			relPath := cg.makeRelativeToTarget(p, target.Directory)
			paths[relPath] = true
		}
	}

	// If no include directory found but headers exist in project, add them
	if len(paths) == 0 && len(cg.Scanner.HeaderFiles) > 0 {
		// Check if headers are in a specific directory
		for _, header := range cg.Scanner.HeaderFiles {
			headerDir := filepath.Dir(header)
			if headerDir != "." && headerDir != "" {
				// Check if header is related to this target
				if target.Directory == "." || strings.HasPrefix(headerDir, target.Directory) {
					relPath := cg.makeRelativeToTarget(headerDir, target.Directory)
					if relPath != "." {
						paths[relPath] = true
					}
				}
			}
		}
	}

	result := []string{}
	for p := range paths {
		result = append(result, p)
	}
	return result
}

// isLibraryUsedByTarget checks if a vendored library is used by the target
func (cg *ConfigGenerator) isLibraryUsedByTarget(target BuildTarget, lib VendoredLibrary) bool {
	// Check if any source file in the target includes headers from this library
	for _, srcFile := range target.SourceFiles {
		if includes, ok := cg.Scanner.IncludeMap[srcFile]; ok {
			for _, inc := range includes {
				for _, libHeader := range lib.HeaderFiles {
					if strings.Contains(inc, filepath.Base(libHeader)) {
						return true
					}
				}
			}
		}
	}
	return false
}

// getExternalLibsForTarget gets external libraries used by a target
func (cg *ConfigGenerator) getExternalLibsForTarget(target BuildTarget) []ExternalLibrary {
	libMap := make(map[string]ExternalLibrary)

	// Collect all includes from target sources
	for _, srcFile := range target.SourceFiles {
		if includes, ok := cg.Scanner.IncludeMap[srcFile]; ok {
			for _, inc := range includes {
				// Check against external libraries
				for _, extLib := range cg.Scanner.ExternalLibs {
					if inc == extLib.HeaderName || strings.Contains(inc, extLib.HeaderName) {
						libMap[extLib.Name] = extLib
					}
				}
			}
		}
	}

	// Convert map to slice
	result := []ExternalLibrary{}
	for _, lib := range libMap {
		result = append(result, lib)
	}
	return result
}

// collectAllIncludes collects all unique includes for documentation
func (cg *ConfigGenerator) collectAllIncludes(target BuildTarget) []string {
	includeMap := make(map[string]bool)

	for _, srcFile := range target.SourceFiles {
		if includes, ok := cg.Scanner.IncludeMap[srcFile]; ok {
			for _, inc := range includes {
				includeMap[inc] = true
			}
		}
	}

	// Sort into categories
	standardIncs := []string{}
	projectIncs := []string{}
	externalIncs := []string{}

	for inc := range includeMap {
		if isStandardHeader(inc) {
			standardIncs = append(standardIncs, inc)
		} else if cg.Scanner.isProjectHeader(inc) {
			projectIncs = append(projectIncs, inc)
		} else {
			externalIncs = append(externalIncs, inc)
		}
	}

	// Combine in order
	result := []string{}
	result = append(result, standardIncs...)
	result = append(result, externalIncs...)
	result = append(result, projectIncs...)

	return result
}

// makeRelativeToTarget makes a path relative to the target directory
func (cg *ConfigGenerator) makeRelativeToTarget(path, targetDir string) string {
	if targetDir == "." || targetDir == "" {
		return path
	}

	// If path is already relative to target, return as-is
	if !strings.HasPrefix(path, targetDir+"/") {
		// Path is in a different directory tree
		if strings.HasPrefix(targetDir, path) {
			// Target is deeper than the path
			rel, _ := filepath.Rel(targetDir, path)
			return rel
		}
		return path
	}

	// Remove target directory prefix
	rel, _ := filepath.Rel(targetDir, path)
	return rel
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func exists(path string) bool {
	_, err := filepath.Glob(path)
	return err == nil
}
