package main

import (
	"fmt"
	"log"

	"github.com/Sabique-Islam/catalyst/internal/fetch"
)

func main() {
	fmt.Println("==============================================")
	fmt.Println("  Catalyst Dependency Scanner Demo           ")
	fmt.Println("==============================================")
	fmt.Println()

	// Scan the entire project recursively
	fmt.Println("Recursively scanning all .c and .h files in: ../../")
	fmt.Println()

	deps, err := fetch.ScanDependencies("../..")
	if err != nil {
		log.Fatalf("Error scanning dependencies: %v", err)
	}

	if len(deps) == 0 {
		fmt.Println("No system dependencies found.")
		return
	}

	fmt.Printf("Found %d unique system dependencies:\n\n", len(deps))
	for i, dep := range deps {
		fmt.Printf("  %d. %s\n", i+1, dep)
	}

	fmt.Println()
	fmt.Println("=============================================")
	fmt.Println("These are abstract package names extracted from:")
	fmt.Println("   #include <package/...> statements")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("   1. Map these to OS-specific package names")
	fmt.Println("   2. Add to catalyst.yml dependencies")
	fmt.Println("   3. Run 'catalyst build' to install them")
}
