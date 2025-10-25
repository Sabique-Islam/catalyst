package pkgdb

import (
	"fmt"
	"strconv"
	"strings"
)

// InteractiveSearch performs a dynamic search and lets the user choose from results
func InteractiveSearch(headerName, pkgManager string) (string, bool) {
	fmt.Printf("Searching for packages that provide '%s' header...\n", headerName)

	results, err := DynamicSearch(headerName, pkgManager)
	if err != nil {
		fmt.Printf("Search failed: %v\n", err)
		return "", false
	}

	if len(results) == 0 {
		fmt.Printf("No packages found for '%s'\n", headerName)
		return "", false
	}

	// If there's only one high-confidence result, use it automatically
	if len(results) == 1 && results[0].Confidence >= 80 {
		fmt.Printf("Found package: %s (confidence: %d%%)\n", results[0].PackageName, results[0].Confidence)
		return results[0].PackageName, true
	}

	// Show options to user
	fmt.Printf("Found %d potential packages for '%s':\n\n", len(results), headerName)

	maxResults := len(results)
	if maxResults > 10 {
		maxResults = 10 // Limit to top 10 results
	}

	for i := 0; i < maxResults; i++ {
		result := results[i]
		fmt.Printf("  %d. %s (confidence: %d%%)\n", i+1, result.PackageName, result.Confidence)
		if result.Description != "" {
			fmt.Printf("     %s\n", result.Description)
		}
		fmt.Println()
	}

	fmt.Printf("  0. Skip this dependency\n\n")

	for {
		fmt.Printf("Choose package (0-%d): ", maxResults)

		var input string
		fmt.Scanln(&input)

		choice, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Println("Please enter a valid number.")
			continue
		}

		if choice == 0 {
			return "", false
		}

		if choice < 1 || choice > maxResults {
			fmt.Printf("Please enter a number between 0 and %d.\n", maxResults)
			continue
		}

		selected := results[choice-1]
		fmt.Printf("Selected: %s\n", selected.PackageName)
		return selected.PackageName, true
	}
}

// BatchSearch performs searches for multiple dependencies with progress indication
func BatchSearch(dependencies []string, pkgManager string, interactive bool) map[string]string {
	results := make(map[string]string)

	fmt.Printf("Resolving %d dependencies for %s...\n\n", len(dependencies), pkgManager)

	for i, dep := range dependencies {
		fmt.Printf("[%d/%d] Processing '%s'...\n", i+1, len(dependencies), dep)

		// Try static translation first
		if pkg, found := Translate(dep, pkgManager); found {
			if pkg != "" { // Skip empty (standard library) packages
				results[dep] = pkg
				fmt.Printf("  ✓ Found in database: %s\n", pkg)
			} else {
				fmt.Printf("  ✓ Standard library header (no package needed)\n")
			}
			fmt.Println()
			continue
		}

		// Try dynamic search
		if interactive {
			if pkg, found := InteractiveSearch(dep, pkgManager); found {
				results[dep] = pkg
			}
		} else {
			if pkg, found := TranslateWithSearch(dep, pkgManager); found {
				results[dep] = pkg
				fmt.Printf("  ✓ Found via search: %s\n", pkg)
			} else {
				fmt.Printf("  ✗ Not found - likely a local header\n")
			}
		}
		fmt.Println()
	}

	return results
}
