package fetch

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// systemIncludeRegex matches system includes and extracts the package name
// Pattern: ^#include <([^\/.]+)(\.h|>|/)
// Captures the first path component before '.', '/', or '>'
var systemIncludeRegex = regexp.MustCompile(`^#include <([^\/.]+)(\.h|>|/)`)

// localIncludeRegex matches local includes and extracts the file name without extension
// Pattern: ^#include "([^"]+)"
// Captures the filename inside quotes
var localIncludeRegex = regexp.MustCompile(`^#include "([^"]+)"`)

// ScanDependencies recursively scans a directory for C/C++ files and extracts
// both system header dependencies from #include <...> and local headers from #include "..."
// It returns a unique list of header names.
func ScanDependencies(rootDir string) ([]string, error) {
	// Use a map as a set to track unique package names
	uniqueDeps := make(map[string]bool)

	// Walk the directory tree
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		// Handle any errors from WalkDir itself
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .c and .h files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".c" && ext != ".h" {
			return nil
		}

		// Process the file
		deps, err := extractDependenciesFromFile(path)
		if err != nil {
			// Log the error but continue processing other files
			fmt.Fprintf(os.Stderr, "Warning: failed to process %s: %v\n", path, err)
			return nil
		}

		// Add to unique set
		for _, dep := range deps {
			uniqueDeps[dep] = true
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory tree: %w", err)
	}

	// Convert map to slice
	result := make([]string, 0, len(uniqueDeps))
	for dep := range uniqueDeps {
		result = append(result, dep)
	}

	return result, nil
}

// extractDependenciesFromFile reads a file line by line and extracts
// both system and local header names from #include statements
func extractDependenciesFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var deps []string
	scanner := bufio.NewScanner(file)

	// Read file line by line
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for system includes: #include <...>
		if strings.HasPrefix(line, "#include <") {
			matches := systemIncludeRegex.FindStringSubmatch(line)
			if len(matches) >= 2 {
				packageName := matches[1]
				deps = append(deps, packageName)
			}
			continue
		}

		// Check for local includes: #include "..."
		if strings.HasPrefix(line, "#include \"") {
			matches := localIncludeRegex.FindStringSubmatch(line)
			if len(matches) >= 2 {
				// Extract filename without path and extension
				fullPath := matches[1]
				fileName := filepath.Base(fullPath)
				// Remove .h extension if present
				fileName = strings.TrimSuffix(fileName, ".h")
				deps = append(deps, fileName)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return deps, nil
}
