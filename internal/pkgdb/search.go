package pkgdb

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// SearchResult represents a search result from a package manager
type SearchResult struct {
	PackageName string
	Description string
	Confidence  int // 0-100, higher is better match
}

// DynamicSearch searches package managers for a dependency when it's not found in the static database
func DynamicSearch(headerName, pkgManager string) ([]SearchResult, error) {
	switch pkgManager {
	case "apt":
		return searchApt(headerName)
	case "dnf":
		return searchDnf(headerName)
	case "pacman":
		return searchPacman(headerName)
	case "brew":
		return searchBrew(headerName)
	case "vcpkg":
		return searchVcpkg(headerName)
	case "choco":
		return searchChoco(headerName)
	default:
		return nil, fmt.Errorf("unsupported package manager: %s", pkgManager)
	}
}

// searchApt searches for packages using apt (Debian/Ubuntu)
func searchApt(headerName string) ([]SearchResult, error) {
	var results []SearchResult

	// First try apt-file to find which package provides the header
	if output, err := exec.Command("apt-file", "search", headerName+".h").Output(); err == nil {
		results = append(results, parseAptFileOutput(string(output), headerName)...)
	}

	// Also try apt search with common variations
	searchTerms := []string{
		headerName,
		"lib" + headerName,
		headerName + "-dev",
		"lib" + headerName + "-dev",
	}

	for _, term := range searchTerms {
		if output, err := exec.Command("apt", "search", term).Output(); err == nil {
			results = append(results, parseAptSearchOutput(string(output), headerName)...)
		}
	}

	return deduplicateResults(results), nil
}

// searchDnf searches for packages using dnf (Fedora/RHEL)
func searchDnf(headerName string) ([]SearchResult, error) {
	var results []SearchResult

	searchTerms := []string{
		headerName,
		headerName + "-devel",
		"lib" + headerName + "-devel",
	}

	for _, term := range searchTerms {
		if output, err := exec.Command("dnf", "search", term).Output(); err == nil {
			results = append(results, parseDnfOutput(string(output), headerName)...)
		}
	}

	return deduplicateResults(results), nil
}

// searchPacman searches for packages using pacman (Arch Linux)
func searchPacman(headerName string) ([]SearchResult, error) {
	var results []SearchResult

	searchTerms := []string{
		headerName,
		"lib" + headerName,
	}

	for _, term := range searchTerms {
		if output, err := exec.Command("pacman", "-Ss", term).Output(); err == nil {
			results = append(results, parsePacmanOutput(string(output), headerName)...)
		}
	}

	return deduplicateResults(results), nil
}

// searchBrew searches for packages using brew (macOS Homebrew)
func searchBrew(headerName string) ([]SearchResult, error) {
	var results []SearchResult

	searchTerms := []string{
		headerName,
		"lib" + headerName,
	}

	for _, term := range searchTerms {
		if output, err := exec.Command("brew", "search", term).Output(); err == nil {
			results = append(results, parseBrewOutput(string(output), headerName)...)
		}
	}

	return deduplicateResults(results), nil
}

// searchVcpkg searches for packages using vcpkg (Windows)
func searchVcpkg(headerName string) ([]SearchResult, error) {
	var results []SearchResult

	if output, err := exec.Command("vcpkg", "search", headerName).Output(); err == nil {
		results = parseVcpkgOutput(string(output), headerName)
	}

	return deduplicateResults(results), nil
}

// searchChoco searches for packages using chocolatey (Windows)
func searchChoco(headerName string) ([]SearchResult, error) {
	var results []SearchResult

	if output, err := exec.Command("choco", "search", headerName).Output(); err == nil {
		results = parseChocoOutput(string(output), headerName)
	}

	return deduplicateResults(results), nil
}

// parseAptFileOutput parses apt-file output to find package names
func parseAptFileOutput(output, headerName string) []SearchResult {
	var results []SearchResult

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// apt-file output format: package: /path/to/file
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			pkgName := strings.TrimSpace(parts[0])
			filePath := strings.TrimSpace(parts[1])

			// Calculate confidence based on file path match
			confidence := calculatePathConfidence(filePath, headerName)
			if confidence > 0 {
				results = append(results, SearchResult{
					PackageName: pkgName,
					Description: fmt.Sprintf("Provides %s", filePath),
					Confidence:  confidence,
				})
			}
		}
	}

	return results
}

// parseAptSearchOutput parses apt search output
func parseAptSearchOutput(output, headerName string) []SearchResult {
	var results []SearchResult

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "WARNING") {
			continue
		}

		// apt search output format: package/suite [arch] version
		parts := strings.Fields(line)
		if len(parts) > 0 {
			pkgName := strings.Split(parts[0], "/")[0]
			confidence := calculateNameConfidence(pkgName, headerName)
			if confidence > 20 { // Only include reasonable matches
				results = append(results, SearchResult{
					PackageName: pkgName,
					Description: strings.Join(parts[1:], " "),
					Confidence:  confidence,
				})
			}
		}
	}

	return results
}

// parseDnfOutput parses dnf search output
func parseDnfOutput(output, headerName string) []SearchResult {
	var results []SearchResult

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "===") {
			continue
		}

		// dnf search output format: package.arch : description
		if strings.Contains(line, " : ") {
			parts := strings.SplitN(line, " : ", 2)
			if len(parts) == 2 {
				pkgName := strings.Split(parts[0], ".")[0]
				confidence := calculateNameConfidence(pkgName, headerName)
				if confidence > 20 {
					results = append(results, SearchResult{
						PackageName: pkgName,
						Description: parts[1],
						Confidence:  confidence,
					})
				}
			}
		}
	}

	return results
}

// parsePacmanOutput parses pacman search output
func parsePacmanOutput(output, headerName string) []SearchResult {
	var results []SearchResult

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// pacman output format: repo/package version
		parts := strings.Fields(line)
		if len(parts) > 0 {
			pkgName := strings.Split(parts[0], "/")
			if len(pkgName) > 1 {
				name := pkgName[1]
				confidence := calculateNameConfidence(name, headerName)
				if confidence > 20 {
					results = append(results, SearchResult{
						PackageName: name,
						Description: strings.Join(parts[1:], " "),
						Confidence:  confidence,
					})
				}
			}
		}
	}

	return results
}

// parseBrewOutput parses brew search output
func parseBrewOutput(output, headerName string) []SearchResult {
	var results []SearchResult

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		confidence := calculateNameConfidence(line, headerName)
		if confidence > 20 {
			results = append(results, SearchResult{
				PackageName: line,
				Description: "Homebrew formula",
				Confidence:  confidence,
			})
		}
	}

	return results
}

// parseVcpkgOutput parses vcpkg search output
func parseVcpkgOutput(output, headerName string) []SearchResult {
	var results []SearchResult

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) > 0 {
			confidence := calculateNameConfidence(parts[0], headerName)
			if confidence > 20 {
				results = append(results, SearchResult{
					PackageName: parts[0],
					Description: strings.Join(parts[1:], " "),
					Confidence:  confidence,
				})
			}
		}
	}

	return results
}

// parseChocoOutput parses chocolatey search output
func parseChocoOutput(output, headerName string) []SearchResult {
	var results []SearchResult

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) > 0 {
			confidence := calculateNameConfidence(parts[0], headerName)
			if confidence > 20 {
				results = append(results, SearchResult{
					PackageName: parts[0],
					Description: strings.Join(parts[1:], " "),
					Confidence:  confidence,
				})
			}
		}
	}

	return results
}

// calculateNameConfidence calculates how well a package name matches the header name
func calculateNameConfidence(pkgName, headerName string) int {
	pkgLower := strings.ToLower(pkgName)
	headerLower := strings.ToLower(headerName)

	// Exact match
	if pkgLower == headerLower {
		return 100
	}

	// Contains header name
	if strings.Contains(pkgLower, headerLower) {
		return 80
	}

	// Header name contains package name
	if strings.Contains(headerLower, pkgLower) {
		return 70
	}

	// Common library naming patterns
	patterns := []string{
		"lib" + headerLower,
		headerLower + "-dev",
		headerLower + "-devel",
		"lib" + headerLower + "-dev",
		"lib" + headerLower + "-devel",
	}

	for _, pattern := range patterns {
		if pkgLower == pattern {
			return 90
		}
		if strings.Contains(pkgLower, pattern) {
			return 60
		}
	}

	// Fuzzy matching (simple edit distance approximation)
	if len(pkgLower) > 0 && len(headerLower) > 0 {
		minLen := len(pkgLower)
		if len(headerLower) < minLen {
			minLen = len(headerLower)
		}

		matches := 0
		for i := 0; i < minLen; i++ {
			if pkgLower[i] == headerLower[i] {
				matches++
			}
		}

		similarity := (matches * 100) / minLen
		if similarity > 60 {
			return similarity / 2 // Reduce confidence for fuzzy matches
		}
	}

	return 0
}

// calculatePathConfidence calculates confidence based on file path matching
func calculatePathConfidence(filePath, headerName string) int {
	pathLower := strings.ToLower(filePath)
	headerLower := strings.ToLower(headerName)

	// Check if the file is actually a header file
	if !strings.HasSuffix(pathLower, ".h") && !strings.HasSuffix(pathLower, ".hpp") {
		return 0
	}

	// Extract filename from path
	fileName := strings.ToLower(filepath.Base(filePath))

	// Exact match
	if fileName == headerLower+".h" || fileName == headerLower+".hpp" {
		return 95
	}

	// Check if it's in a reasonable include path
	includePatterns := []string{"/usr/include/", "/usr/local/include/", "/opt/include/"}
	for _, pattern := range includePatterns {
		if strings.Contains(pathLower, pattern) {
			if strings.Contains(fileName, headerLower) {
				return 80
			}
		}
	}

	// General path contains header name
	if strings.Contains(fileName, headerLower) {
		return 60
	}

	return 0
}

// deduplicateResults removes duplicate search results and sorts by confidence
func deduplicateResults(results []SearchResult) []SearchResult {
	seen := make(map[string]SearchResult)

	// Keep the result with highest confidence for each package
	for _, result := range results {
		existing, exists := seen[result.PackageName]
		if !exists || result.Confidence > existing.Confidence {
			seen[result.PackageName] = result
		}
	}

	// Convert back to slice and sort by confidence (highest first)
	var deduplicated []SearchResult
	for _, result := range seen {
		deduplicated = append(deduplicated, result)
	}

	// Simple sort by confidence (bubble sort for simplicity)
	for i := 0; i < len(deduplicated)-1; i++ {
		for j := 0; j < len(deduplicated)-i-1; j++ {
			if deduplicated[j].Confidence < deduplicated[j+1].Confidence {
				deduplicated[j], deduplicated[j+1] = deduplicated[j+1], deduplicated[j]
			}
		}
	}

	return deduplicated
}

// GetBestMatch returns the best matching package from search results
func GetBestMatch(results []SearchResult) (string, bool) {
	if len(results) == 0 {
		return "", false
	}

	best := results[0]
	return best.PackageName, best.Confidence >= 50 // Only return if confidence is reasonable
}
