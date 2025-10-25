package fetch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// SymbolInfo represents information about undefined symbols
type SymbolInfo struct {
	Symbol string
	File   string
	Type   string // "function", "variable", etc.
}

// MissingDependency represents a missing implementation
type MissingDependency struct {
	Symbols        []SymbolInfo
	SuggestedFiles []string
	SuggestedLibs  []string
	PossibleCauses []string
	Category       string
}

// ScanMissingSymbols attempts to compile and detect missing symbols
func ScanMissingSymbols(projectPath string) ([]MissingDependency, error) {
	// Find all C source files
	sourceFiles, err := findSourceFiles(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find source files: %w", err)
	}

	if len(sourceFiles) == 0 {
		return nil, nil // No source files to analyze
	}

	// Try linking directly to catch undefined symbols
	linkArgs := append(sourceFiles, "-o", "/tmp/catalyst_test_link")
	cmd := exec.Command("gcc", linkArgs...)
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()

	// Always clean up test files
	os.Remove("/tmp/catalyst_test_link")

	if err == nil {
		return nil, nil // No missing symbols
	}

	// Parse linker output for undefined references
	return parseLinkErrors(string(output))
}

// findSourceFiles locates all C/C++ source files in the project
func findSourceFiles(projectPath string) ([]string, error) {
	var sources []string

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip hidden and build directories, but not the root directory
			name := filepath.Base(path)
			if (strings.HasPrefix(name, ".") && path != projectPath) || name == "build" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check for C/C++ source files
		ext := filepath.Ext(path)

		if ext == ".c" || ext == ".cpp" || ext == ".cc" || ext == ".cxx" {
			relPath, err := filepath.Rel(projectPath, path)
			if err != nil {
				return err
			}
			sources = append(sources, relPath)
		}

		return nil
	})

	return sources, err
}

// parseLinkErrors parses compiler/linker output for undefined symbols
func parseLinkErrors(output string) ([]MissingDependency, error) {
	var dependencies []MissingDependency

	// Regex to match: "undefined reference to `function_name'"
	re := regexp.MustCompile(`undefined reference to \x60([^']+)'`)
	matches := re.FindAllStringSubmatch(output, -1)

	if len(matches) == 0 {
		return nil, nil
	}

	// Group symbols by category
	symbols := make(map[string][]string)
	for _, match := range matches {
		if len(match) > 1 {
			symbol := match[1]
			category := categorizeSymbol(symbol)
			symbols[category] = append(symbols[category], symbol)
		}
	}

	// Remove duplicates and generate suggestions
	for category, symbolList := range symbols {
		// Remove duplicates
		uniqueSymbols := removeDuplicateStrings(symbolList)

		dep := MissingDependency{
			Symbols:  convertToSymbolInfo(uniqueSymbols),
			Category: category,
		}

		// Generate suggestions based on symbol category
		generateSuggestions(&dep, category)

		dependencies = append(dependencies, dep)
	}

	return dependencies, nil
}

// categorizeSymbol determines the category of a missing symbol
func categorizeSymbol(symbol string) string {
	symbol = strings.ToLower(symbol)

	if strings.Contains(symbol, "print") || strings.Contains(symbol, "color") || strings.Contains(symbol, "terminal") {
		return "print"
	}
	if strings.Contains(symbol, "hash") || strings.Contains(symbol, "map") {
		return "hashmap"
	}
	if strings.Contains(symbol, "embedding") || strings.Contains(symbol, "vector") || strings.Contains(symbol, "ml") {
		return "embedding"
	}
	if strings.Contains(symbol, "activity") || strings.Contains(symbol, "recommend") {
		return "activity"
	}
	if strings.Contains(symbol, "json") || strings.Contains(symbol, "parse") {
		return "json"
	}
	if strings.Contains(symbol, "math") || strings.Contains(symbol, "sqrt") || strings.Contains(symbol, "pow") {
		return "math"
	}
	if strings.Contains(symbol, "thread") || strings.Contains(symbol, "pthread") || strings.Contains(symbol, "mutex") {
		return "threading"
	}
	if strings.Contains(symbol, "curl") || strings.Contains(symbol, "http") || strings.Contains(symbol, "net") {
		return "network"
	}
	if strings.Contains(symbol, "file") || strings.Contains(symbol, "read") || strings.Contains(symbol, "write") {
		return "fileio"
	}

	// Extract base name for generic categorization
	parts := strings.Split(symbol, "_")
	if len(parts) > 0 {
		return parts[0]
	}

	return "misc"
}

// generateSuggestions creates suggestions based on symbol category
func generateSuggestions(dep *MissingDependency, category string) {
	switch category {
	case "print":
		dep.SuggestedFiles = []string{"utils.c", "print.c", "terminal.c", "colors.c"}
		dep.SuggestedLibs = []string{"ncurses", "termcap"}
		dep.PossibleCauses = []string{
			"Missing implementation file for printing functions",
			"Need to link terminal/color libraries",
			"Missing utility functions implementation",
		}

	case "hashmap":
		dep.SuggestedFiles = []string{"hashmap.c", "data_structures.c", "hash.c"}
		dep.SuggestedLibs = []string{"glib-2.0"}
		dep.PossibleCauses = []string{
			"Missing hashmap implementation file",
			"Need custom data structure library",
			"Consider using GLib hash tables",
		}

	case "embedding":
		dep.SuggestedFiles = []string{"embeddings.c", "ml.c", "vectors.c", "neural.c"}
		dep.SuggestedLibs = []string{"blas", "lapack", "openblas", "cblas"}
		dep.PossibleCauses = []string{
			"Missing machine learning implementation",
			"Need linear algebra library",
			"Missing vector computation functions",
		}

	case "json":
		dep.SuggestedFiles = []string{"json.c", "parser.c"}
		dep.SuggestedLibs = []string{"jansson", "json-c", "cjson"}
		dep.PossibleCauses = []string{
			"Missing JSON parsing library",
			"Need to install JSON library",
			"Missing custom JSON implementation",
		}

	case "math":
		dep.SuggestedLibs = []string{"m", "gsl", "fftw3"}
		dep.PossibleCauses = []string{
			"Need to link math library (-lm)",
			"Missing advanced math library",
		}

	case "threading":
		dep.SuggestedLibs = []string{"pthread"}
		dep.PossibleCauses = []string{
			"Need to link pthread library (-lpthread)",
			"Missing threading implementation",
		}

	case "network":
		dep.SuggestedFiles = []string{"network.c", "http.c", "client.c"}
		dep.SuggestedLibs = []string{"curl", "libcurl"}
		dep.PossibleCauses = []string{
			"Missing network implementation",
			"Need HTTP client library",
		}

	case "activity":
		dep.SuggestedFiles = []string{"activities.c", "recommender.c", "engine.c"}
		dep.PossibleCauses = []string{
			"Missing core application logic implementation",
			"Need to implement recommendation engine",
		}

	default:
		dep.SuggestedFiles = []string{fmt.Sprintf("%s.c", category)}
		dep.PossibleCauses = []string{
			fmt.Sprintf("Missing implementation file for %s functions", category),
		}
	}
}

// convertToSymbolInfo converts string symbols to SymbolInfo structs
func convertToSymbolInfo(symbols []string) []SymbolInfo {
	var result []SymbolInfo
	for _, symbol := range symbols {
		result = append(result, SymbolInfo{
			Symbol: symbol,
			Type:   "function", // Default assumption
		})
	}
	return result
}

// removeDuplicateStrings removes duplicate strings from a slice
func removeDuplicateStrings(input []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range input {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// ExtractSymbolNames extracts symbol names from SymbolInfo slice
func ExtractSymbolNames(symbols []SymbolInfo) []string {
	var names []string
	for _, sym := range symbols {
		names = append(names, sym.Symbol)
	}
	return names
}
