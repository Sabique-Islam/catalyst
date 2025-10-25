package build

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Compile compiles the Go project into a binary in bin/
func Compile() error {
	fmt.Println("Compiling project...")

	// Determine output binary path
	binaryPath := "bin/project"
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	// Ensure bin directory exists
	if err := os.MkdirAll("bin", 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Run go build
	cmd := exec.Command("go", "build", "-o", binaryPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	fmt.Printf("✅ Compilation complete: %s\n", binaryPath)
	return nil
}
